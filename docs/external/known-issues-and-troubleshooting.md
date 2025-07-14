# Known issues and troubleshooting

## Firefox

### MOZILLA_PKIX_ERROR_KEY_PINNING_FAILURE

It's possible when visiting sites, that the page will fail to render and you will see

"Secure Connection Failed" with the Error code: `MOZILLA_PKIX_ERROR_KEY_PINNING_FAILURE`

A work around to this is setting `security.cert_pinning.enforcement_level` in `about:config` to 1

![alt text](firefox.png)

If you're using arkenfox you can add the above to your `user-overrides.js` like so

```js
user_pref("security.cert_pinning.enforcement_level", 1);
```

### MOZILLA_PKIX_ERROR_MITM_DETECTED

Another issue that may pop up is `MOZILLA_PKIX_ERROR_MITM_DETECTED`

This can be fixed in a few ways:

Sometimes a simple restart of the browser is all that's needed, if that doesn't work you can go to

`Settings > Privacy & Security` and scroll down to `Certificates`

Once there, tick the checkbox for "Allow Firefox to automatically trust third-party root certificates you install"

And finally you can set `security.enterprise_roots.enabled` to `true` in `about:config` and restart Firefox. This appears to have the same effect as the above fix, according to the Firefox docs. However, it's good to check both just incase.
