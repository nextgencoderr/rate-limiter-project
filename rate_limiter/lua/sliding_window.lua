-- Atomic sliding window counter: KEYS[1]=window key
-- ARGV[1]=window_ms, ARGV[2]=limit, ARGV[3]=now_ms, ARGV[4]=member (unique request id)
local key = KEYS[1]
local window_ms = tonumber(ARGV[1])
local limit = tonumber(ARGV[2])
local now_ms = tonumber(ARGV[3])
local member = ARGV[4]

local window_start = now_ms - window_ms
redis.call('ZREMRANGEBYSCORE', key, '-inf', window_start)
local count = redis.call('ZCARD', key)

local allowed = 0
if count < limit then
  redis.call('ZADD', key, now_ms, member)
  allowed = 1
  count = count + 1
end

redis.call('PEXPIRE', key, window_ms + 1000)
return {allowed, limit - count}
