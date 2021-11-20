# commitment.ninja

This is just a silly little commitment scheme to prove that you knew
something before you could disclose it. It appends some entropy to the
end so that the recipient can't guess what you've committed.

## Usage

```
❯ curl -X POST https://commitment.ninja -d 'I know a thing!'
message: I know a thing! 6f422054f19f9294f11924e190339728
commit: 7d3dff1515112b4c0fe3c0c0395348e40c423d96dd1b0d9354580027a7c88419

❯ echo -n "I know a thing! 6f422054f19f9294f11924e190339728" | shasum -a 256
7d3dff1515112b4c0fe3c0c0395348e40c423d96dd1b0d9354580027a7c88419  -
```

The response body is YAML to make it human-readable. Share the commit
and save the message (with the entropy!) until you're ready to reveal
it.
