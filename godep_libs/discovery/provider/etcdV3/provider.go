package etcdV3

import (
	"context"
	"fmt"
	"time"

	"github.com/coreos/etcd/clientv3"
	"motify_core_api/godep_libs/discovery"
	"motify_core_api/godep_libs/discovery/provider"
)

const (
	defaultKeyTTL = 10 * time.Second
)

// Timeout constants are placed in "var" section to redefine the values in Unit Tests.
var (
	// minimum time before registration retries
	registerRetryInterval = 1 * time.Second

	// watch retry timeout
	watchRetryAfter = 1 * time.Second

	// progressReportTimeout is a timeout to fail Watch if no ProgressNotify events occur.
	// It's slightly bigger than 20 minutes, because etcd sends ProgressNotify every 10 minutes
	// if no event occurs. But if there was an event - it postpones the next send.
	// https://godoc.org/github.com/coreos/etcd/clientv3#WithProgressNotify
	// So the maximum gap between 2 notifications is 20 minutes!
	// And we need to have some reserve for network delivery delay.
	progressReportTimeout = 20*time.Minute + 30*time.Second
	// watchCreateTimeout is timeout for recieving "Created" event from Watch channel
	watchCreateTimeout = 30 * time.Second

	// defaultGetTimeout is default timeout for provider.Get() operation
	defaultGetTimeout = 10 * time.Second
	// defaultRegisterTimeout is default timeout for key registration
	defaultRegisterTimeout = 10 * time.Second
	// defaultRevokeTimeout is default timeout for lease revoke
	defaultRevokeTimeout = 10 * time.Second
)

type providerEtcd struct {
	logger  discovery.ILogger
	client  *clientv3.Client
	kvAPI   clientv3.KV
	watcher clientv3.Watcher
}

// NewProvider creates IProvider updating service information in etcd
func NewProvider(client *clientv3.Client, logger discovery.ILogger) provider.IProvider {
	if logger == nil {
		logger = discovery.NewNilLogger()
	}

	return &providerEtcd{
		logger:  logger,
		client:  client,
		kvAPI:   clientv3.NewKV(client),
		watcher: clientv3.NewWatcher(client),
	}
}

// Get returns all values matching the given key filter
func (p *providerEtcd) Get(ctx context.Context, filter provider.KeyFilter) ([]provider.KV, error) {
	if _, ok := ctx.Deadline(); !ok {
		timeoutCtx, cancel := context.WithTimeout(ctx, defaultGetTimeout)
		ctx = timeoutCtx
		defer cancel()
	}

	prefix := keyPrefix(filter)
	res, err := p.kvAPI.Get(ctx, prefix, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		return nil, err
	}

	kvs := make([]provider.KV, 0, len(res.Kvs))
	for _, etcdKV := range res.Kvs {
		kv, err := discoveryKV(etcdKV)
		if err != nil {
			return nil, err
		}
		kvs = append(kvs, kv)
	}

	return kvs, nil
}

// RegisterValues validates registration data, runs keepalive loop and blocks until context is finished.
func (p *providerEtcd) RegisterValues(ctx context.Context, kvs ...provider.KV) error {
	if err := validateRegistrationData(kvs...); err != nil {
		return err
	}

	defer p.logger.Debugf("register loop exit")
	for {
		minInterval := time.After(registerRetryInterval)

		leaseID, err := p.doRegister(ctx, kvs...)
		select {
		case <-ctx.Done():
			p.revoke(leaseID)
			return nil
		default:
		}
		p.logger.Warningf("could not register values: %s", err)

		<-minInterval
	}
}

// Watch watches for value updates and sends them to results channel.
// All existing values are also sent to prevent any data loss between getting existing values
// and starting watching the updates.
func (p *providerEtcd) Watch(ctx context.Context, filter provider.KeyFilter) <-chan *provider.Event {
	prefix := keyPrefix(filter)

	r := make(chan *provider.Event)
	go func() {
		p.logger.Debugf("watching for %q by prefix", prefix)
		defer p.logger.Debugf("done watching %q", prefix)

		currentRevision := p.fetchExisting(ctx, prefix, r)
		var err error
		for {
			currentRevision, err = p.doWatch(ctx, prefix, currentRevision, r)
			select {
			case <-ctx.Done():
				close(r)
				return
			default:
			}
			p.logger.Warningf("%q client watch failed, current revision: %d, error: %s", prefix, currentRevision, err)

			<-time.After(watchRetryAfter)
		}
	}()

	return r
}

// fetchExisting gets existing values and send put events to out channel.
// Max modified revision is returned for further use as watching start point.
func (p *providerEtcd) fetchExisting(ctx context.Context, prefix string, out chan<- *provider.Event) int64 {
	p.logger.Debugf("%q fetching existing values", prefix)

	for {
		timeoutCtx, cancel := context.WithTimeout(ctx, defaultGetTimeout)
		defer cancel()
		res, err := p.kvAPI.Get(timeoutCtx, prefix, clientv3.WithPrefix())
		if err == nil {
			maxRevision := res.Header.Revision
			if len(res.Kvs) == 0 {
				p.logger.Debugf("%q no existing values found", prefix)
				return maxRevision
			}

			event, err := newEventFromKVs(res.Kvs)
			if err != nil {
				p.logger.Warningf("%q failed to create event: %s", prefix, err)
				return maxRevision
			}

			p.sendEvent(ctx, event, prefix, out)
			return maxRevision
		}

		p.logger.Warningf("%q failed to fetch existing values: %s", prefix, err)
		select {
		case <-ctx.Done():
			return 0
		case <-time.After(watchRetryAfter):
			p.logger.Debugf("%q retry fetching", prefix)
		}
	}
}

// doWatch processes watch events for given key prefix.
// Last modified revision is returned for further use as watching start point.
func (p *providerEtcd) doWatch(ctx context.Context, prefix string, revision int64, out chan<- *provider.Event) (int64, error) {
	p.logger.Debugf("init %q watch client, revision: %d", prefix, revision)
	defer p.logger.Debugf("finish watching %q", prefix)

	ctx, cancel := context.WithCancel(ctx)
	ctx = clientv3.WithRequireLeader(ctx)
	defer cancel()

	watchCh, err := p.createWatchChan(ctx, prefix, revision)
	if err != nil {
		return revision, err
	}

	// We need this ugly timers to check if Watch is stuck because of network problems.
	// Maybe in some bright future it will be handled inside etcd client itself:
	// https://github.com/coreos/etcd/issues/7247
	timer := time.NewTimer(progressReportTimeout)
	defer timer.Stop()
	createdTimer := time.NewTimer(watchCreateTimeout)
	defer createdTimer.Stop()

	lastRevision := revision
	for {
		select {
		case resp, ok := <-watchCh:
			if !ok {
				return lastRevision, fmt.Errorf("Watch channel closed unexpectedly")
			}

			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(progressReportTimeout)
			if resp.Canceled || resp.Err() != nil {
				// TODO: switch back to checking "Canceled" flag.
				// For now it's buggy, and any error causes channel close
				// See https://github.com/coreos/etcd/issues/8231
				return lastRevision, resp.Err()
			}

			if resp.Created {
				p.logger.Debugf("%q watcher successfully created", prefix)
				if !createdTimer.Stop() {
					<-createdTimer.C
				}
			}
			if resp.IsProgressNotify() {
				p.logger.Debugf("Progress notify, rev: %d", lastRevision)
			}
			if resp.Header.Revision > lastRevision {
				lastRevision = resp.Header.Revision
			}
			for _, event := range resp.Events {
				p.processWatchEvent(ctx, event, prefix, out)
			}
		case <-timer.C:
			return lastRevision, fmt.Errorf("ProgressNotify timeout")
		case <-createdTimer.C:
			return lastRevision, fmt.Errorf("CreatedNotify timeout")
		}
	}

	return lastRevision, fmt.Errorf("WTF, this code should never be reached")
}

// createWatchChan initializes clientv3.WatchChan with proper creation timeout.
//
// watcher.Watch may hang, see https://github.com/coreos/etcd/issues/7247
// So we should run it in separate goroutine and return the channel when it's ready
func (p *providerEtcd) createWatchChan(ctx context.Context, prefix string, revision int64) (clientv3.WatchChan, error) {
	timer := time.NewTimer(watchCreateTimeout)
	defer timer.Stop()

	res := make(chan clientv3.WatchChan, 1)
	go func() {
		// start watch from the next revision, otherwise you will get a duplicate event
		c := p.watcher.Watch(ctx, prefix, clientv3.WithPrefix(), clientv3.WithRev(revision+1),
			clientv3.WithProgressNotify(), clientv3.WithCreatedNotify())
		res <- c
		close(res)
	}()

	select {
	case watchChan := <-res:
		return watchChan, nil
	case <-timer.C:
		return nil, fmt.Errorf("failed to initialize Watch client")
	}
}

// processWatchEvent converts etcd3 Event into provider.Event and sends to the channel
func (p *providerEtcd) processWatchEvent(ctx context.Context, event *clientv3.Event, prefix string, out chan<- *provider.Event) {
	res, err := watchResponse(event)
	if err != nil {
		p.logger.Warningf("event parse error: %s", err)
		return
	}

	p.logger.Debugf("%q watcher: %s %s", prefix, res.Type, event.Kv.Key)
	p.sendEvent(ctx, res, prefix, out)
}

// sendEvent send event to the channel
func (p *providerEtcd) sendEvent(ctx context.Context, event *provider.Event, prefix string, out chan<- *provider.Event) {
	select {
	case <-ctx.Done():
		// don't want to block on channel write in case of done context
		p.logger.Debugf("%q context is Done, ", prefix)
	case out <- event:
	}
}

// doRegister runs all put operations with Lease TTL and waits on lease keepalive channel
func (p *providerEtcd) doRegister(ctx context.Context, kvs ...provider.KV) (clientv3.LeaseID, error) {
	// Yeah, we need to create new Leaser every time, because it may corrupt its state.
	// See https://github.com/coreos/etcd/issues/7472
	// We may recieve ErrKeepAliveHalted at leaser.KeepAlive()
	// https: //godoc.org/github.com/coreos/etcd/clientv3#ErrKeepAliveHalted
	// But instead of reinitializing some common p.leaser with lock let's have a leaser per active registration.
	// The code is much easier and clearer.
	leaser := clientv3.NewLease(p.client)
	defer leaser.Close()

	timeoutCtx, cancel := context.WithTimeout(ctx, defaultRegisterTimeout)
	defer cancel()

	resp, err := leaser.Grant(timeoutCtx, int64(defaultKeyTTL.Seconds()))
	if err != nil {
		return 0, fmt.Errorf("lease grant error: %s", err)
	}
	leaseID := resp.ID

	ops := p.prepareOps(leaseID, kvs...)
	if _, err := p.kvAPI.Txn(timeoutCtx).Then(ops...).Commit(); err != nil {
		return 0, fmt.Errorf("lease %d transaction error: %s", leaseID, err)
	}

	keepAlives, err := leaser.KeepAlive(ctx, leaseID)
	if err != nil {
		// almost unreal situation, but better to handle it - revoke the lease
		go p.revoke(leaseID)
		return 0, fmt.Errorf("lease %d keep alive error: %s", leaseID, err)
	}

	// wait on keepAlives channel
	for range keepAlives {
	}

	// it's not really an error if ctx is canceled - this situation should be handled in the caller
	return leaseID, fmt.Errorf("keepalive timeout, TTL reached")
}

// revoke revokes the lease, all the keys are deleted at the same moment
func (p *providerEtcd) revoke(leaseID clientv3.LeaseID) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultRevokeTimeout)
	defer cancel()
	if _, err := p.client.Revoke(ctx, leaseID); err != nil {
		p.logger.Warningf("can't revoke leaseID: %d", leaseID)
	}
}

// prepareOps prepare etcd operations for transaction, setting lease for TTL update
func (p *providerEtcd) prepareOps(leaseID clientv3.LeaseID, kvs ...provider.KV) []clientv3.Op {
	ops := make([]clientv3.Op, 0, len(kvs))

	for _, kv := range kvs {
		key := kv.RawKey
		if key == "" {
			// if RawKey is provided we just use it, otherwise we build service key
			key = newServiceKey(kv.Namespace, kv.Service).storageKey()
		}
		ops = append(ops, clientv3.OpPut(key, kv.Value, clientv3.WithLease(leaseID)))
	}

	return ops
}
