local bw = require('bw')
local mod = {}

local new = function(uri)
    local obj = {
        uri = uri or "",
        _state = nil,
        _last_commit = nil,
        _switch_mode = nil,
        _switch_name = nil
    }
    print("Switch at", uri)
    bw.subscribe(uri .. "/signal/info", "2.1.1.2", function(uri, msg)
        obj._state = msg["state"]
        obj._last_commit = msg["last_commit"]
        obj._switch_mode = msg["switch_mode"]
        obj._switch_name = msg["switch_name"]
    end)
    return setmetatable(obj, mod)
end

local state = function(self, val)
    if val == nil then
        return self._state
    end
    bw.publish(self.uri .. "/slot/state", "2.1.1.2", {state=val})
    return self._state
end
mod.state = state

local last_commit = function(self)
    return self._last_commit
end
mod.last_commit = last_commit

local switch_name = function(self)
    return self._switch_name
end
mod.switch_name = switch_name

local switch_mode = function(self)
    return self._switch_mode
end
mod.switch_mode = switch_mode

mod.__index = mod
local ctor = function(cls, ...)
    return new(...)
end

return setmetatable({}, {__call = ctor})
