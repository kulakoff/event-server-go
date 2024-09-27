
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
```
2024-09-25T22:20:59 Opening door by RFID 00000075BC01AD, apartment 0
```

##### RFID external reader
```
2024-09-27 07:55:30 Opening door by external RFID 000000C69798DA, apartment 0
```


##### Open door by button
```
2024-09-27 08:05:43	Main door button pressed!
2024-09-27 08:05:43	Main door opened by button press
2024-09-27 08:05:46	Main door button unpressed!
2024-09-27 08:05:50	Main door button pressed!
2024-09-27 08:05:50	Main door opened by button press
```

##### Open door by button. Additional door
```
2024-09-27 08:06:33	Additional door button pressed!
2024-09-27 08:06:33	Alt door opened by button press
2024-09-27 08:06:35	Additional door button unpressed!
```

##### Motion start
```
2024-09-27 07:24:00	SS_MAINAPI_ReportAlarmHappen(0, 2)
```

##### Motion stop
```
2024-09-27 07:24:03	SS_MAINAPI_ReportAlarmFinish(0, 2)
```

##### other
```
2024-09-26 05:39:02	IMX122_AutoIRcut_Task to Day ----------
2024-09-26 18:52:20	IMX122_AutoIRcut_Task to Night ----------
2024-09-26 19:51:15	IMX122_AutoIRcut_Task to Day ----------
2024-09-26 21:21:40	IMX122_AutoIRcut_Task to Night ----------
```