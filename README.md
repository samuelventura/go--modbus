# go-modbus

- [x] Master
- [x] TCP Protocol
- [x] RTU Protocol
- [x] TCP Transport
- [x] Connect and Read timeout
- [x] Discard before send
- [x] Test: Address and function sweept
- [x] Exception custom error
- [x] Serial Transport: separated package
- [x] Model: Basic map model
- [x] Disable trace in production
- [x] Slave: Bootstrap only
- [x] Environ controlled trace
- [ ] Out of bounds checks
- [ ] Special function codes
- [ ] Special data types
- [ ] Verify and narrow public api
- [ ] Test: IO error recoveries
- [ ] Test: In the middle breaks

## Notes

- Responses don't have enough info to be fully parsed.
A ReadDis for example wont packet count but only
the total bytes making it dependant on knowing 
the original intended count.
