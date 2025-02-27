---
title: Application Gateway
---

Application Gateway is an intermediary component between a Function or a service and an external API. 
It [proxies the requests](../../../05-technical-reference/00-architecture/ac-03-application-gateway.md) from Functions and services in Kyma to external APIs based on the configuration stored in Secrets.

## Supported API authentication for Application CR

Application Gateway can call services which are not secured, or are secured with:

- [Basic Authentication](https://tools.ietf.org/html/rfc7617)
- OAuth
- Client certificates

Additionally, Application Gateway supports cross-site request forgery (CSRF) tokens as an optional layer of API protection.

Application Gateway calls the registered APIs accordingly, basing on the security type specified for the API in the Application CR.

## Provide a custom access token

Application Gateway overrides the registered API's security type if it gets a request which contains the **Access-Token** header. In such a case, Application Gateway rewrites the token from the **Access-Token** header into an OAuth-compliant **Authorization** header and forwards it to the target API.

This mechanism is suited for implementations in which an external application handles user authentication.

See how to [pass an access token in a request header](../../../04-operation-guides/operations/ac-01-pass-access-token-in-request-header.md).