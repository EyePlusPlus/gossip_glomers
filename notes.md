Expt: type handler is misnamed
result: 
No handler for {"id":2,"src":"c2","dest":"n0","body":{"echo":"Please echo 108","type":"echo","msg_id":1}}
Full STDERR logs are available in /Users/mansishah/personal/recurse/gossip-glomers/maelstrom/store/echo/20250903T164858.510-0400/node-logs/n0.log


Expt: Respond with a bad type value
result: clojure.lang.ExceptionInfo: Malformed RPC response. Maelstrom sent node n0 the following request:

Expt: Return bad value for read
result: Empty list in the response

Solution for Broadcast challenge was to also check for float value and it worked.
