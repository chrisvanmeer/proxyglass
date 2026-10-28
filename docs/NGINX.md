# nginx reverse proxy

Proxyglass serves plain HTTP and is intended to sit behind nginx for TLS. Bind its listener to
loopback (`127.0.0.1:8080`) and preserve the public host/scheme headers for OIDC redirects.

```nginx
location / {
    proxy_pass http://127.0.0.1:8080;
    proxy_http_version 1.1;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

Set the OIDC redirect URI to `https://<public-host>/ui/oidc/callback` when configuring a fixed
callback. If `oidc.redirect_uri` is empty, Proxyglass derives the callback from the incoming Host and
`X-Forwarded-Proto` headers.
