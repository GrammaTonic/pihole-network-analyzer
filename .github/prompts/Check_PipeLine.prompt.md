---
mode: agent
---

Check the latest pipeline run for the repository and return the status of the run.
If the pipeline is running, wair for it to finish with a loop script, and check If the pipeline is successful, return the status as "success". If the pipeline fails, return the status as "failure" then investigate the cause of the failure and fx the issue.
Commit your changes and push them to the repository so that the pipeline can run again. This is the loop until the pipeline is successful.
