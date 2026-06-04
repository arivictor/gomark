---
title: Security
description: Best practices for securing your GoMark site and runner, including authentication, authorization, and sandboxing.
order: 4
---

## CSRF

GoMark includes built-in protections against Cross-Site Request Forgery (CSRF) attacks. GoMark automatically generates a unique CSRF token for each user session. 

This token is included in the HTML of your docs pages and must be sent with any request to the runner's code execution endpoint. This ensures that only requests originating from your docs site can execute code on the runner, preventing malicious sites from tricking users into executing unwanted code.

## Runner

The runner is a powerful component that executes Go code snippets from your docs. To secure it, you should:

1. Use authentication: GoMark supports several authentication modes for the runner, including static bearer tokens. Choose the one that best fits your security requirements and ensure that only authorized users can access the runner.

2. Limit execution: Configure the runner to set limits on execution time and resource usage to prevent abuse and ensure that it remains responsive.

3. Keep it updated: Regularly update GoMark to benefit from security patches and improvements.