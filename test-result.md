env NETGEAR_SWITCHES="tswitch1=None1234@" ./build/examples/switch-test tswitch1
Switch Test Program - ntgrrc Library Validation
===============================================

Connecting to switch: tswitch1
✓ Authentication successful (Model: GS308EPP)
✓ Detected 8 POE ports, 0 ethernet ports

POE Power Cycling Test:
  Port 1: ✗ POE still delivering power after disable
  Port 2: ✗ POE still delivering power after disable
  Port 3: ✓ Disabled → ✓ Verified Off → ✓ Enabled → ✓ Verified On
  Port 4: ✓ Disabled → ✓ Verified Off → ✓ Enabled → ✓ Verified On
  Port 5: ✓ Disabled → ✓ Verified Off → ✓ Enabled → ✓ Verified On
  Port 6: ✗ POE still delivering power after disable
  Port 7: ✗ POE still delivering power after disable
  Port 8: ✓ Disabled → ✓ Verified Off → ✓ Enabled → ✓ Verified On

Bandwidth Limitation Test:
  ⚠ No ethernet ports detected, skipping bandwidth test

LED Control Test:
  ⚠ LED control not supported in current library implementation
  ✓ LEDs disabled → ✓ LEDs enabled → ✓ State restored

Final Validation:
✓ All settings restored to original state

Test Summary:
  Total Operations: 11
  Successful: 6
  Failed: 5
  Duration: 1m6s

Test Results:
  ✗ poe_cycling (failed on ports: [1 2 6 7]) - Error: POE cycling failed on 4 ports
  ✓ bandwidth_limiting (no ethernet ports)
  ✓ led_control (LED control test completed)

⚠ 5 operations failed out of 11 total
bs:~/herlein/src/ntgrrc>
