be aware  that the following code was ommitted from this chart:

```yaml
### Namespace ###
kind: Namespace
apiVersion: v1
metadata:
  name: {{.Namespace}}
  {{- if and .EnableTLS .ProxyAutoInjectEnabled }}
  labels:
    {{.ProxyAutoInjectLabel}}: disabled
  {{- end }}
```

because Helm doesn't currently support namespace templates.

This needs to be manually implemented in supergloo when
support for linkerd sidecar injection is added.


TODO:
also need to add the templates for proxy injector when ready