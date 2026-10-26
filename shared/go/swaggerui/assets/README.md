# Bundled Swagger UI

These files are copied from the official `swagger-ui-dist` npm package,
version `5.33.1`, published by the Swagger API project:

- `swagger-ui-bundle.js`
- `swagger-ui.css`
- `LICENSE`, `NOTICE`, and `swagger-ui-bundle.js.LICENSE.txt`

The distribution is licensed under Apache-2.0. The shared Go package embeds
the bundle into File Service, Moments, and Stream Service binaries. File Service
serves it only from authenticated `/f/docs/...` routes when
`api_docs.enabled=true`. Each documentation page loads its UI and OpenAPI
document from the same LiteNAS origin; it makes no runtime CDN request.

When updating the bundle, copy the matching distribution files and license
notices together, update the version in `../swaggerui.go`, and verify each page
still loads its local OpenAPI document.
