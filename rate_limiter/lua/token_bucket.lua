-- Atomic token bucket: KEYS[1]=bucket key
-- ARGV[1]=capacity, ARGV[2]=refill_rate (tokens/sec), ARGV[3]=now_ms, ARGV[4]=cost (default 1)
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2])
local now_ms = tonumber(ARGV[3])
local cost = tonumber(ARGV[4]) or 1

local data = redis.call('HMGET', key, 'tokens', 'last_refill_ms')
local tokens = tonumber(data[1])
local last_refill = tonumber(data[2])

if tokens == nil then
  tokens = capacity
  last_refill = now_ms
else
  local elapsed_sec = math.max(0, (now_ms - last_refill) / 1000.0)
  tokens = math.min(capacity, tokens + elapsed_sec * refill_rate)
  last_refill = now_ms
end

local allowed = 0
local remaining = tokens
if tokens >= cost then
  tokens = tokens - cost
  allowed = 1
  remaining = tokens
end

redis.call('HMSET', key, 'tokens', tokens, 'last_refill_ms', last_refill)
redis.call('PEXPIRE', key, math.ceil((capacity / refill_rate) * 2000))

return {allowed, remaining}
