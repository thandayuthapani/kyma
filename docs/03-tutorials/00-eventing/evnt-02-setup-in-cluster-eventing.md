---
title: Publish events without Kyma Application
---

In-cluster Eventing allows publishers to send messages and subscribers to receive them without the need for a Kyma Application. This means that instead of the usual event flow where Application Connector publishes events to the Event Publisher Proxy, events can be published from within the cluster directly to the Event Publisher Proxy.

1. To use in-cluster Eventing, create a subscription where the **eventType.value** field includes the name of your application. In the following example, this is `sap.kyma.custom.nonexistingapp.order.created.v1`, where `nonexistingapp` is an application that doesn't exist in Kyma.

```yaml
apiVersion: eventing.kyma-project.io/v1alpha1
kind: Subscription
metadata:
  name: mysub
  namespace: default
spec:
  filter:
    filters:
    - eventSource:
        property: source
        type: exact
        value: ""
      eventType:
        property: type
        type: exact
        value: sap.kyma.custom.nonexistingapp.order.created.v1
  sink: http://myservice.default.svc.cluster.local
```

2. On the publisher side, include the exact same Application name in the `type` field, like in the following example:

```yaml
curl -k -i \
    --data @<(cat <<EOF
    {
        "source": "kyma",
        "specversion": "1.0",
        "eventtypeversion": "v1",
        "data": {"orderCode":"3211213"},
        "datacontenttype": "application/json",
        "id": "759815c3-b142-48f2-bf18-c6502dc0998f",
        "type": "sap.kyma.custom.nonexistingapp.order.created.v1"
    }
EOF
    ) \
    -H "Content-Type: application/cloudevents+json" \
    "http://eventing-event-publisher-proxy.kyma-system/publish"
```

> **NOTE:** If you want to use a Function to publish a CloudEvent, see the [Event object SDK specification](../../05-technical-reference/svls-08-function-specification.md#event-object-sdk).
