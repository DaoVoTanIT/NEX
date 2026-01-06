#!/usr/bin/env bash
set -euo pipefail

#######################################
# CONFIG
#######################################
SENTINEL_CONTAINER="redis-sentinel-1"
SENTINEL_PORT=26379
MASTER_NAME="mymaster"

REDIS_PORT=6379
REDIS_PASSWORD="${REDIS_PASSWORD:-}"
TEST_KEY="ha:test"
WAIT_TIMEOUT=40

#######################################
# UTILS
#######################################
ts() { date +"[%Y-%m-%d %H:%M:%S]"; }
log() { echo "$(ts) $1"; }
fail() { echo "$(ts) ❌ $1"; exit 1; }

#######################################
# PRECHECK
#######################################
[[ -z "$REDIS_PASSWORD" ]] && fail "REDIS_PASSWORD is not set"
docker ps >/dev/null || fail "Docker not running"

#######################################
# SENTINEL
#######################################
get_master_ip() {
  docker exec "$SENTINEL_CONTAINER" redis-cli -p "$SENTINEL_PORT" \
    SENTINEL get-master-addr-by-name "$MASTER_NAME" | sed -n '1p'
}

#######################################
# REDIS OPS
#######################################
redis_write() {
  local host="$1"
  docker exec "$SENTINEL_CONTAINER" redis-cli \
    -h "$host" -p "$REDIS_PORT" -a "$REDIS_PASSWORD" \
    SET "$TEST_KEY" "$(date)" | grep -q OK \
    || fail "Redis write failed"
}

#######################################
# FIND MASTER CONTAINER
#######################################
find_master_container() {
  for c in $(docker ps --format '{{.Names}}' | grep redis); do
    if docker exec "$c" redis-cli -a "$REDIS_PASSWORD" INFO replication 2>/dev/null \
      | grep -q "role:master"; then
      echo "$c"
      return
    fi
  done
  fail "Cannot find Redis master container"
}

#######################################
# WAIT FAILOVER
#######################################
wait_failover() {
  local old="$1"
  log "Waiting for failover..."

  for ((i=1;i<=WAIT_TIMEOUT;i++)); do
    new="$(get_master_ip)"
    if [[ -n "$new" && "$new" != "$old" ]]; then
      log "New master: $new"
      return
    fi
    sleep 1
  done

  fail "Failover timeout"
}

#######################################
# VERIFY REJOIN
#######################################
verify_rejoin() {
  local container="$1"
  log "Verify old master rejoined as replica"

  sleep 3
  docker exec "$container" redis-cli -a "$REDIS_PASSWORD" INFO replication \
    | grep -q "role:slave" \
    || fail "Old master did NOT rejoin as replica"

  log "Old master rejoined OK"
}

#######################################
# MAIN
#######################################
log "Discover current master"
OLD_MASTER_IP="$(get_master_ip)"
[[ -z "$OLD_MASTER_IP" ]] && fail "Cannot detect master"

log "Current master: $OLD_MASTER_IP"

log "Write BEFORE failover"
redis_write "$OLD_MASTER_IP"

MASTER_CONTAINER="$(find_master_container)"
log "Stopping master container: $MASTER_CONTAINER"
docker stop "$MASTER_CONTAINER" >/dev/null

wait_failover "$OLD_MASTER_IP"

NEW_MASTER_IP="$(get_master_ip)"

log "Write AFTER failover"
redis_write "$NEW_MASTER_IP"

log "Restart old master"
docker start "$MASTER_CONTAINER" >/dev/null

verify_rejoin "$MASTER_CONTAINER"

log "✅ REDIS SENTINEL FULL HA TEST PASSED 🎉"
