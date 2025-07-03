# Known issues and troubleshooting

### Firefox

It's possible when visiting sites, that the page will fail to render and you will see

"Secure Connection Failed" with the Error code: `MOZILLA_PKIX_ERROR_KEY_PINNING_FAILURE`

A work around to this is setting `security.cert_pinning.enforcement_level` in `about:config` to 1

![alt text](firefox.png)

If you're using arkenfox you can add the above to your `user-overrides.js` like so

```js
user_pref("security.cert_pinning.enforcement_level", 1);
```
