Expt: type handler is misnamed
result: 
No handler for {"id":2,"src":"c2","dest":"n0","body":{"echo":"Please echo 108","type":"echo","msg_id":1}}
Full STDERR logs are available in /Users/mansishah/personal/recurse/gossip-glomers/maelstrom/store/echo/20250903T164858.510-0400/node-logs/n0.log


Expt: Respond with a bad type value
result: clojure.lang.ExceptionInfo: Malformed RPC response. Maelstrom sent node n0 the following request:

Expt: Return bad value for read
result: Empty list in the response

Solution for Broadcast challenge was to also check for float value and it worked.

------
challenge 4
1) Is global counter common to all the nodes?
2) Will multiple nodes receive the add operation?
3) 




2025-09-04 15:34:34 EDT
Managed to get the code mostly written out. The kv operations are bonking.
2025-09-04 15:40:26 EDT Add logs for read message and more nicer error string logging in the file
- Read operations are failing with key not found error.
- Solution: Assume 0 if keynotexists error occurs and proceed with cas

2025-09-04 17:26:29 EDT
- Race conditions seem to be happening, can't store the values simultaneously. But could use some retry logic.
- Can also try keeping local counter and only using global counter to read values for other nodes. When would the write happen though?
- Approach 1:
    - Use node-wise key in the kv store to store the values. 
    - Read from all nodes to get the combined values.
    - Use topology handler to track other nodes


2025-09-05 13:10:50 EDT
- No logs showing up so far
- No handler for read. Resolved
- Remove excessive logs in read handler
- Everything works. Need to check Eval script matches
- Challenge 4 completed
- send msg not getting parsed correctly. Resolved now
- Figured out how to see the application code logs without writing to tmp file
- Is offset unique for each k key?
- Completed the challenge. Main hurdle was handling concurrent reads and writes to a map object (multithreaded programming). Using Mutex to solve it, and looks like it worked


