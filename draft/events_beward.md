
##### call example flow
```json
{"time":"2024-09-24T23:08:12.58074843+03:00","level":"INFO","msg":"Processing Beward message","srcIP":"37.235.143.99","host":"192.168.13.152","message":"Emulating call to apartment 1(0)!"}
{"time":"2024-09-24T23:08:12.9544609+03:00","level":"INFO","msg":"Processing Beward message","srcIP":"37.235.143.99","host":"192.168.13.152","message":"[61255] Calling sip:1000000020@rbt-demo.lanta.me:50142 through account 0(0)..."}
{"time":"2024-09-24T23:08:13.020613835+03:00","level":"INFO","msg":"Processing Beward message","srcIP":"37.235.143.99","host":"192.168.13.152","message":"[61255] SIP call 4 state changed to CALLING"}
{"time":"2024-09-24T23:08:22.633260021+03:00","level":"INFO","msg":"Processing Beward message","srcIP":"37.235.143.99","host":"192.168.13.152","message":"[61255] SIP call 4 state changed to EARLY"}
{"time":"2024-09-24T23:08:30.363943993+03:00","level":"INFO","msg":"Processing Beward message","srcIP":"37.235.143.99","host":"192.168.13.152","message":"[61255] SIP call 4 state changed to CONNECTING"}
{"time":"2024-09-24T23:08:31.928974327+03:00","level":"INFO","msg":"Processing Beward message","srcIP":"37.235.143.99","host":"192.168.13.152","message":"Opened device HisiAudDev for capture, sample rate=8000, ch=1, bits=16, latency=100 ms"}
{"time":"2024-09-24T23:08:31.969668767+03:00","level":"INFO","msg":"Processing Beward message","srcIP":"37.235.143.99","host":"192.168.13.152","message":"[61255] SIP call 4 state changed to CONFIRMED"}
{"time":"2024-09-24T23:08:32.329120025+03:00","level":"INFO","msg":"Processing Beward message","srcIP":"37.235.143.99","host":"192.168.13.152","message":"Incoming DTMF RFC2833 on call 4: 1"}
{"time":"2024-09-24T23:08:32.351315585+03:00","level":"INFO","msg":"Processing Beward message","srcIP":"37.235.143.99","host":"192.168.13.152","message":"[61255] Opening door by DTMF command for apartment 1"}
{"time":"2024-09-24T23:08:32.631757092+03:00","level":"INFO","msg":"Processing Beward message","srcIP":"37.235.143.99","host":"192.168.13.152","message":"[61255] SIP talk started for apartment 1"}
{"time":"2024-09-24T23:08:32.784658359+03:00","level":"INFO","msg":"Processing Beward message","srcIP":"37.235.143.99","host":"192.168.13.152","message":"Incoming DTMF RFC2833 on call 4: 1"}
{"time":"2024-09-24T23:08:32.798607844+03:00","level":"INFO","msg":"Processing Beward message","srcIP":"37.235.143.99","host":"192.168.13.152","message":"[61255] Opening door by DTMF command for apartment 1"}
{"time":"2024-09-24T23:08:33.286974793+03:00","level":"INFO","msg":"Processing Beward message","srcIP":"37.235.143.99","host":"192.168.13.152","message":"Incoming DTMF RFC2833 on call 4: 1"}
{"time":"2024-09-24T23:08:33.299818603+03:00","level":"INFO","msg":"Processing Beward message","srcIP":"37.235.143.99","host":"192.168.13.152","message":"[61255] Opening door by DTMF command for apartment 1"}
{"time":"2024-09-24T23:08:35.848249538+03:00","level":"INFO","msg":"Processing Beward message","srcIP":"37.235.143.99","host":"192.168.13.152","message":"[61255] SIP call 4 is DISCONNECTED [reason=200 (Normal call clearing)]"}
{"time":"2024-09-24T23:08:38.649189204+03:00","level":"INFO","msg":"Processing Beward message","srcIP":"37.235.143.99","host":"192.168.13.152","message":"[61255] SIP call done for apartment 1, handset is down"}
{"time":"2024-09-24T23:08:40.5515295+03:00","level":"INFO","msg":"Processing Beward message","srcIP":"37.235.143.99","host":"192.168.13.152","message":"[61255] All calls are done for apartment 1"}
```

##### RFID
```json
{"time":"2024-09-25T22:20:59.550499106+03:00","level":"INFO","msg":"Processing Beward message","srcIP":"192.168.88.25","host":"localhost.localdomain","message":"Opening door by RFID 00000075BC01AD, apartment 0"}
```