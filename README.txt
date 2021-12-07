
[ ] Master
[ ] TCP Protocol
[ ] RTU Protocol
[ ] TCP Transport
[ ] Connect and Read TO
[ ] Discard before send
[ ] Serial Transport
[ ] Testing
[ ] Out of bounds checks
[ ] Exception parser
[ ] Slave
[ ] Model
[ ] Special function codes
[ ] Special data types
[ ] Disable trace in production
[ ] Verify and narrow public api

Notice

- Responses have not enough info to be fully parsed
A ReadDis for example wont packet count but only
the total bytes making it dependant on knowing 
the original intended count.


Fixme

- Simplify SetError/Discard/Finally
