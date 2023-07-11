# commitment.ninja

This is just a silly little commitment scheme to prove that you knew
something before you could disclose it.

## Usage

```
❯ curl -X POST https://commitment.ninja -d 'I know a thing!'
message: I know a thing!
key: a94e1d5df17e0b9a517f4086ed554e737a158c64cc7e1a67d56aec24fca758e8
commit: bbdea5ff72b3336a7476fe204183bf1da4c8317d5f2b6b36adc33b9d39845caa

❯ key=a94e1d5df17e0b9a517f4086ed554e737a158c64cc7e1a67d56aec24fca758e8
❯ echo -n 'I know a thing!' | openssl sha256 -mac HMAC -macopt "hexkey:${key}"
SHA2-256(stdin)= bbdea5ff72b3336a7476fe204183bf1da4c8317d5f2b6b36adc33b9d39845caa
```

The response body is YAML to make it human-readable. Share the commit
and save the key and message until you're ready to reveal it.
