# Problems we hit and how we fixed them

Each entry: what broke, what we saw, the cause, the fix, the lesson.

<!-- template
## Short title
- **Symptom:**
- **Cause:**
- **Fix:**
- **Lesson:**
-->



## Containers could not reach each other in Codespaces
- **Symptom:** kafka-init failed with "Timed out waiting for a node assignment."
  The same command worked from inside the Kafka container.
- **Cause:** Codespaces runs Docker inside a container, with two firewall systems
  installed: iptables (nft) and iptables-legacy. Docker wrote its allow rules to the
  nft one. The legacy one still had `FORWARD DROP` as its default, so all
  container-to-container traffic was silently dropped.
- **How we found it:** name lookup worked but TCP connections hung (hang = dropped,
  refused = nothing listening). Every container pair failed, not just Kafka, so it was
  the network, not Kafka. The iptables warning pointed at the legacy tables.
- **Fix:** `sudo iptables-legacy -P FORWARD ACCEPT` (dev machine only; resets when the
  Codespace stops).
- **Lesson:** the error message pointed at Kafka, but the cause was the network layer.
  Test the layer below before changing config.