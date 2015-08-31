#!/bin/bash

# absolute path to executable binary
progpath='/usr/sbin/incus'
confpath='/etc/incus'
logpath='/var/log/incus.log'

# binary program name
prog=$(basename $progpath)

# pid file
pidfile="/var/run/${prog}.pid"

# make sure full path to executable binary and conf path is found
! [ -x $progpath ] && echo "$progpath: executable not found" && exit 1
! [ -f "$confpath/config.yml" ] && echo "$confpath/config.yml: configuration file not found" && exit 1

ulimit -n 1000000

eval_cmd() {
  local rc=$1
  if [ $rc -eq 0 ]; then
    echo '[  OK  ]'
  else
    echo '[FAILED]'
  fi
  return $rc
}

start() {
  # see if running
  local pids=$(pidof $prog)

  export GOTRACEBACK=1

  if [ -n "$pids" ]; then
    echo "$prog (pid $pids) is already running"
    return 0
  fi
  printf "%-50s%s" "Starting $prog: " ''
  $progpath --conf="$confpath/"  >> $logpath 2>&1 & 

  # save pid to file if you want
  echo $! > $pidfile

  # check again if running
  pidof $prog >/dev/null 2>&1
  eval_cmd $?
}

stop() {
  # see if running
  local pids=$(pidof $prog)

  if [ -z "$pids" ]; then
    echo "$prog not running"
    return 0
  fi
  printf "%-50s%s" "Stopping $prog: " ''
  rm -f $pidfile
  kill $pids
  eval_cmd $?
}

status() {
  # see if running
  local pids=$(pidof $prog)

  if [ -n "$pids" ]; then
    echo "$prog (pid $pids) is running"
  else
    echo "$prog is stopped"
  fi
}

case $1 in
  start)
    start
    ;;
  stop)
    stop
    ;;
  status)
    status
    ;;
  restart)
    stop
    sleep 1
    start
    ;;
  *)
    echo "Usage: $0 {start|stop|status|restart}"
    exit 1
esac

exit $?

