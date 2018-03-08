FROM centos

ADD simple* ./

CMD sleep 2; \
    [ -f /var/log/syslog ] && cat /dev/null > /var/log/syslog; \
        cat /dev/null > /var/log/out.log; \
        cat /dev/null > /var/log/err.log; \
    for i in `seq 1 10`; do \
        ./simple -service "example_service_$i" -network "unixgram" -address "/var/log/socket" \
            -rt "10s" -pause "100ms" -c 2000 -buffer 32000000 -workers 4 >> /var/log/out.log 2>> /var/log/err.log & done; \
    wait; \
    echo -n "simple err: "; cat /var/log/err.log | wc -l; \
    echo -n "simple out: "; cat /var/log/out.log | wc -l; \
    echo -n "simple syslog: "; cat /var/log/syslog | wc -l; \
    [ -f ./simple_old ] && ( \
        for i in `seq 1 10`; do \
           ./simple_old -service "example_service_old_$i" -network "unixgram" -address "/var/log/socket" \
               -rt "10s" -pause "100ms" -c 2000 >> /var/log/out.log 2>> /var/log/err.log & done; \
        wait; \
        echo ""; \
        echo -n "simple_old err: "; cat /var/log/err.log | grep "example_service_old_" | wc -l; \
        echo -n "simple_old out: "; cat /var/log/out.log | grep "example_service_old_" | wc -l; \
        echo -n "simple_old syslog: "; cat /var/log/syslog  | grep "example_service_old_" | wc -l; \
    )