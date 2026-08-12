# timeful

## Project setup

```
npm install
```

### Compiles and hot-reloads for development

```
npm run serve
```

### Compiles and minifies for production

```
npm run build
```

### Runtime configuration

Browser configuration is read from `window.configs`, defined by
`public/config.js`. The combined production image built from the repository
root generates that file at container startup from `VUE_APP_POSTHOG_API_KEY`,
`VUE_APP_GOOGLE_CLIENT_ID`, and `VUE_APP_MICROSOFT_CLIENT_ID`. Edit
`public/config.js` directly for local development. `frontend/Dockerfile` remains
a standalone development build/export image.

### Customize configuration

See [Configuration Reference](https://cli.vuejs.org/config/).
