-- KEYS[1] = oldKey (auth:refresh:<oldJti>)
-- KEYS[2] = reuseKey (auth:reuse:<oldJti>)
-- KEYS[3] = newKey (auth:refresh:<newJti>)
-- ARGV[1] = expectUserId
-- ARGV[2] = newTtlSeconds (number)

-- return:
-- 0: old not exists
-- -1: user with old not match
-- 2: old used (reuse flag)
-- 1: rotated successfully

local oldVal = redis.call("GET", KEYS[1])
if not oldVal then
    return 0
end

-- check if the user id in the old value matches the expected user id
if oldVal["uid"] ~= ARGV[1] then 
    return -1
end

-- check if the old value has been reused
if redis.call("EXISTS", KEYS[2]) == 1 then
    return 2
end

-- mark old jti as used
redis.call("SET", KEYS[2], 1)

-- delete the old value
redis.call("DEL", KEYS[1])

-- write new refresh jti with TTL
redis.call("SET", KEYS[3], ARGV[1], "EX", tonumber(ARGV[2]))

return 1