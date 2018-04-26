package dconfig

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	etcdcl "github.com/coreos/etcd/client"
	"github.com/jpillora/backoff"
)

const maxHistoryLen = 10

var keyRE = regexp.MustCompile("^[a-zA-Z0-9_-]+$")

type Manager struct {
	serviceID    string
	ns           string
	venture      string
	env          string
	settingsKey  string
	keysAPI      etcdcl.KeysAPI
	etcdReceiver chan *etcdcl.Response
	mu           sync.RWMutex
	subs         []Subscriber
	logger       Logger
}

type Subscriber func(*Setting)

type Logger interface {
	Error(context.Context, string, ...interface{})
	Notice(context.Context, string, ...interface{})
	Warning(context.Context, string, ...interface{})
}

func NewManager(serviceID string, logger Logger) *Manager {
	return &Manager{
		serviceID:    serviceID,
		etcdReceiver: make(chan *etcdcl.Response, 1),
		logger:       logger,
	}
}

func (m *Manager) Run(client etcdcl.Client, ns, venture, env string) {
	m.ns = ns
	m.venture = venture
	m.env = env
	m.settingsKey = "/" + strings.Join([]string{m.ns, m.venture, m.env, m.serviceID, "settings"}, "/")
	m.keysAPI = etcdcl.NewKeysAPI(client)

	ctx := context.Background()
	go func() {
		defer close(m.etcdReceiver)

		// TODO: what if we have network error here?
		settings, err := m.GetSettings(ctx)
		if err != nil {
			if err == context.DeadlineExceeded || err == context.Canceled {
				return
			}
			m.logger.Error(nil, err.Error())
		}

		knownVars := map[string]struct{}{}
		for _, setting := range settings {
			if v, ok := globalConf.get(setting.Key); ok {
				if err := globalConf.set(setting.Key, setting.Value); err != nil {
					m.logger.Error(nil, err.Error())
				}
				if v.Description != setting.Description {
					if _, err := m.keysAPI.Set(ctx, m.settingsKey+"/"+setting.Key+
						"/description", v.Description, &etcdcl.SetOptions{}); err != nil {
						m.logger.Error(nil, err.Error())
					}
				}
				knownVars[setting.Key] = struct{}{}
			}
		}
		type value struct {
			value       string
			description string
		}
		pendingValues := map[string]value{}
		globalConf.visit(func(name string, v *Var) {
			if _, alreadyRegistered := knownVars[name]; !alreadyRegistered {
				pendingValues[name] = value{
					value:       v.Value.String(),
					description: v.Description,
				}
			}
		})
		for name, v := range pendingValues {
			// owner value is empty because this setting is registered by server not user
			err := m.registerSetting(ctx, "", name, v.value, v.description)
			if err != nil {
				if err == context.DeadlineExceeded || err == context.Canceled {
					return
				}
				// hide error 105: Key already exists (/settings)
				if etcdErr, ok := err.(*etcdcl.Error); ok {
					if etcdErr.Code != etcdcl.ErrorCodeNodeExist {
						m.logger.Error(nil, err.Error())
					}
				} else {
					m.logger.Error(nil, err.Error())
				}
			}
		}

		backoff := backoff.Backoff{}
		for {
			var waitIndex uint64 = 0
			watcher := m.keysAPI.Watcher(m.settingsKey, &etcdcl.WatcherOptions{
				AfterIndex: waitIndex,
				Recursive:  true,
			})
		WATCH:
			for {
				resp, err := watcher.Next(ctx)
				switch err {
				case nil:
					waitIndex = resp.Node.ModifiedIndex
					m.etcdReceiver <- resp
					backoff.Reset()
				case context.DeadlineExceeded, context.Canceled:
					return
				default:
					m.logger.Error(nil, err.Error())
					// Prevent errors from consuming all resources.
					select {
					case <-time.After(backoff.Duration()):
						break WATCH
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	go func() {
		for {
			response, ok := <-m.etcdReceiver
			if !ok {
				return
			}
			if err := m.handleEtcdResponse(ctx, response); err != nil {
				if err == context.DeadlineExceeded || err == context.Canceled {
					return
				}
				m.logger.Error(nil, err.Error())
			}
		}
	}()
}

func (m *Manager) Subscribe(sub Subscriber) {
	m.mu.Lock()
	m.subs = append(m.subs, sub)
	m.mu.Unlock()
}

func (m *Manager) handleEtcdResponse(ctx context.Context, response *etcdcl.Response) error {
	if response.Action != "create" {
		return nil
	}
	prop := strings.TrimPrefix(response.Node.Key, m.settingsKey)
	if prop == "" {
		return nil
	}
	levels := strings.Split(prop, "/")
	if len(levels) < 5 || levels[4] != "value" {
		return nil
	}

	propKey := m.settingsKey + "/" + levels[1]
	resp, err := m.keysAPI.Get(ctx, propKey, &etcdcl.GetOptions{Recursive: true})
	if err != nil {
		return err
	}

	setting := nodeToSetting(resp.Node)

	m.mu.RLock()
	for _, sub := range m.subs {
		go sub(setting)
	}
	m.mu.RUnlock()

	if v, ok := globalConf.get(setting.Key); ok {
		setting.Description = v.Description
	} else {
		//log.Noticef("received unregisterd setting %q with value %q", setting.Key, setting.Value)
		return nil
	}
	if err := globalConf.set(setting.Key, setting.Value); err != nil {
		m.logger.Warning(nil, "could not set setting %q value %q: %v", setting.Key, setting.Value, err)
	}

	return nil
}

func (m *Manager) GetSettings(ctx context.Context) ([]*Setting, error) {
	resp, err := m.keysAPI.Get(ctx, m.settingsKey, &etcdcl.GetOptions{Recursive: true})
	if err != nil {
		return nil, err
	}

	settings := make([]*Setting, resp.Node.Nodes.Len())
	i := 0
	for _, settingNode := range resp.Node.Nodes {
		setting := nodeToSetting(settingNode)
		//v, ok := globalConf.get(setting.Key)
		//if ok {
			// setting.Description = v.Description
			settings[i] = setting
			i++
		//}
	}
	return settings[:i], nil
}

func (m *Manager) GetAllSettings(ctx context.Context) ([]*Setting, error) {
	resp, err := m.keysAPI.Get(ctx, m.settingsKey, &etcdcl.GetOptions{Recursive: true})
	if err != nil {
		return nil, err
	}

	settings := make([]*Setting, resp.Node.Nodes.Len())
	for i, settingNode := range resp.Node.Nodes {
		setting := nodeToSetting(settingNode)
		settings[i] = setting
	}

	return settings, nil
}

func (m *Manager) registerSetting(ctx context.Context, owner, key, value, description string) error {
	if err := checkKey(key); err != nil {
		return err
	}

	settingKey := m.settingsKey + "/" + key

	_, err := m.keysAPI.Set(ctx, settingKey, "", &etcdcl.SetOptions{Dir: true, PrevExist: etcdcl.PrevNoExist})
	if err != nil {
		return err
	}

	_, err = m.keysAPI.Set(ctx, settingKey+"/values", "", &etcdcl.SetOptions{Dir: true, PrevExist: etcdcl.PrevNoExist})
	if err != nil {
		return err
	}

	_, err = m.keysAPI.Create(ctx, settingKey+"/description", description)
	if err != nil {
		return err
	}

	valueKey := settingKey + "/values/" + strconv.FormatInt(time.Now().UTC().UnixNano(), 10)

	_, err = m.keysAPI.Set(ctx, valueKey, "", &etcdcl.SetOptions{Dir: true, PrevExist: etcdcl.PrevNoExist})
	if err != nil {
		return err
	}

	_, err = m.keysAPI.Create(ctx, valueKey+"/owner", owner)
	if err != nil {
		return err
	}

	_, err = m.keysAPI.Create(ctx, valueKey+"/value", value)
	if err != nil {
		return err
	}

	return nil
}

// TODO: i dont like that we have 3 keys for a prop, maybe use json?
func (m *Manager) EditSetting(ctx context.Context, owner, key, newValue string) error {
	if err := checkKey(key); err != nil {
		return err
	}

	settingKey := m.settingsKey + "/" + key

	_, err := m.keysAPI.Get(ctx, settingKey, &etcdcl.GetOptions{})
	if err != nil {
		return err
	}

	if m.purgeStaleHistory(ctx, settingKey); err != nil {
		return err
	}

	valueKey := settingKey + "/values/" + strconv.FormatInt(time.Now().UTC().UnixNano(), 10)

	_, err = m.keysAPI.Set(ctx, valueKey, "", &etcdcl.SetOptions{Dir: true, PrevExist: etcdcl.PrevNoExist})
	if err != nil {
		return err
	}

	_, err = m.keysAPI.Create(ctx, valueKey+"/owner", owner)
	if err != nil {
		return err
	}

	_, err = m.keysAPI.Create(ctx, valueKey+"/value", newValue)
	if err != nil {
		return err
	}

	return nil
}

func (m *Manager) purgeStaleHistory(ctx context.Context, settingKey string) error {
	resp, err := m.keysAPI.Get(ctx, settingKey, &etcdcl.GetOptions{Recursive: true})
	if err != nil {
		return err
	}

	setting := nodeToSetting(resp.Node)
	if len(setting.History) >= maxHistoryLen {
		for i := len(setting.History) - 1; i >= maxHistoryLen-1; i-- {
			snapshot := setting.History[i]
			_, err := m.keysAPI.Delete(ctx, snapshot.key, &etcdcl.DeleteOptions{Dir: true, Recursive: true})
			if err != nil {
				m.logger.Notice(nil, "failed to delete stailed setting history by key %q", snapshot.key, err)
			}
		}
	}
	return nil
}

func checkKey(key string) error {
	if !keyRE.MatchString(key) {
		return errors.New("Invalid key")
	}
	return nil
}
