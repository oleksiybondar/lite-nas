# API Request Architecture

## Purpose

The admin panel uses request builders to keep request execution readable without
spreading response and state branching across providers or feature components.
The builder style is intentional: callers describe an action, attach lifecycle
hooks for side effects, and execute the action at the end of the chain.

This document describes the intended split between generic request lifecycle,
transport-level fetch behavior, and application-level API behavior.

## Design Goals

- Keep React providers thin and focused on exposing context contracts.
- Keep request assembly and execution outside React render flow.
- Avoid repeated `if response.ok` and `try/catch/finally` branching in feature
  code when the branch only exists to update loading or local state.
- Use lifecycle hooks for action side effects such as starting loaders, applying
  response state, resetting state, and ending loaders.
- Keep auth transport behavior in one layer so refresh, retry, cookie handling,
  and login redirect are not duplicated by feature code.
- Allow dependency injection between builders so each layer can be tested in
  isolation.

## Builder Layers

### Request Builder

The request builder is the lifecycle core. It should not know about `fetch`,
React, routing, cookies, or LiteNAS auth.

It owns the common action lifecycle:

- `onBeforeSend`: observe the action immediately before it starts.
- `onSuccess`: handle a successful action result.
- `onError`: handle a failed action result.
- `onEnded`: observe action completion after success, failure, or thrown errors.
- `execute`: run the action and resolve the final result.

`onBeforeSend` and `onEnded` are side-effect hooks. They should not replace the
result. `onSuccess` and `onError` may return a replacement result when the layer
needs to transform what the caller receives.

This layer exists so different action implementations can share the same
lifecycle semantics without copying branching code.

### Fetch Request Builder

The fetch request builder is the transport layer. It should compose or extend
the request lifecycle core and provide fetch-specific request assembly:

- URL
- HTTP method
- headers
- body
- credentials
- direct `fetch` execution

It owns browser transport concerns, including the BFF auth transport behavior:

- include cookies where required
- call `/api/auth/refresh` after protected requests return `401`
- retry the original fetch request after successful refresh
- navigate to `/login` when refresh fails with `401`

The fetch request builder is the only layer that should retry protected requests
after `401`. API-level builders and React providers must not duplicate this
logic.

### API Request Builder

The API request builder is the application-facing BFF layer. It wraps a fetch
request builder and exposes the same lifecycle shape at a higher abstraction.

It should expose feature-friendly methods such as:

- `get(url)`
- `post(url, payload?)`
- `put(url, payload?)`
- `delete(url)`

These methods should return a chainable API request action, not execute
immediately. Feature code can then attach application-level hooks before calling
`execute`.

The shortcut payload argument is only convenience. A request action should also
allow explicit chain-based request customization when that is clearer:

```ts
await api
  .post("/api/example")
  .payload(payload)
  .header("X-Trace-ID", traceID)
  .onSuccess(handleSuccess)
  .execute();
```

The same applies to transport-level request customization where the lower layer
owns the concern. For example, fetch builders may expose headers, body,
credentials, and cookie-backed request behavior, while API builders expose only
the customization that is appropriate for application callers.

The API request builder does not refresh auth, retry `401`, mutate auth state,
or know how cookies are transported. It delegates those details to the fetch
request builder.

## Why Two Hook Layers Are Useful

The fetch layer and API layer both benefit from the same lifecycle shape, but
they represent different abstractions.

At the fetch layer, hooks describe transport behavior around a direct browser
request. This is where BFF-powered fetch behavior lives.

At the API layer, hooks describe a business-facing API action. Feature code
should be able to use the same action lifecycle for UI state:

```ts
const response = await api
  .post("/api/example", payload)
  .onBeforeSend(() => setIsSaving(true))
  .onSuccess(async (response) => {
    const body = await response.clone().json();
    setEntity(body.data);
  })
  .onError(() => {
    setEntity(null);
  })
  .onEnded(() => setIsSaving(false))
  .execute();
```

Using the same lifecycle semantics at both layers is reasonable because the
layers serve different purposes:

- the fetch builder handles direct fetch and BFF transport behavior
- the API builder handles application-level API action behavior

The split becomes a problem only if ownership is blurred. Auth retry belongs to
the fetch builder only. Auth state belongs to auth providers or auth-specific
logic only. API actions should not duplicate transport behavior.

## Unauthorized Redirect Edge Case

The protected API flow has one important edge case.

When the fetch request builder receives `401`, refresh fails with `401`, and the
router navigates to `/login`, the API request builder should not run
application-level `onError` or `onEnded` hooks for that unauthorized response.

The reason is state safety. Once the transport layer has decided the user is no
longer recoverably authenticated, feature-level hooks could update stale page
state while the app is moving to the login route.

The fetch builder should therefore report this case as an internal terminal
outcome to the API builder. The public API can still resolve with `Response`
where appropriate, but the API builder needs enough internal metadata to skip
application-level hooks for redirect-triggering unauthorized outcomes.

## Non-Goals

- The API request builder should not become a query cache.
- The fetch request builder should not own React state.
- Feature components should not manually implement refresh-and-retry behavior.
- Auth providers should not duplicate the protected API `401` retry path.
- Lifecycle hooks should not be used to hide complex business workflows that
  deserve named functions.
