---
name: Bug report
about: Something is broken.
title: ''
labels: ''
assignees: ''

---

**Describe the bug**
A clear and concise description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. terraform plan/apply/refresh '...'
2. See error '...'

**Expected behavior**
A clear and concise description of what you expected to happen.

**Terraform files**
Provide a minimalist main.tf and other terraform files that reproduces your issue.   If your problem is related to terraform plan/apply/etc, most likely these files will be REQUIRED.

**Desktop (please complete the following information):**
 - OS: [e.g. Linux/Mac/Windows Ver 1.2.3]
 - Terraform Version: [e.g. 0.12.9]
 - Ovftool Version:
 - This Plugin Version [e.g. v1.6.0]

**Additional context**
 - Add any other context about the problem here.
 - Enable terraform debugging, then re-run commands that produced the error:
    - export TF_LOG=DEBUG
    - Initially this may not be required, however, be prepared to upload debug information IF REQUESTED.
