# Password of the Day

Sometimes you want to have a password that nobody knows until they
actually need to log in.  That's what this does.

## How it Works

The tool ships with the capability to resolve multiple parts of the
process.  The first half is to generate and set the password.  The
rotation of the password itself is achieved by calling `potd generate
--apply` which will generate a new password, modify `/etc/shadow` to
contain the new password (by default for the root user), and use the
file `/etc/issue.tpl` to rewrite the contents of `/etc/issue`.

Its necessary to rewrite `/etc/issue` each time the password is
changed to get the challenge token.  This token is the thing that
allows a user to obtain the root password.  The password itself is
generated as the output of a seeded generation process seeded with the
hostname, the random challenge token, and a shared value used by the
resolver to define the password domain.

A user would observe a login banner like so:

```
Welcome Traveler!

This system is for authorized use only, and is monitored.  Tread with care.

correct-horse-battery-staple

examplebox login:
```

The XKCD-style string is the "Challenge Token" which is used to allow
a suitably authorized user to recover the root password.  The token,
combined with the hostname, allows the resolver to generate the same
conditions as the password generator and produce the same password.

## Running the Resolver

The password resolver is a web service that allows users to resolve a
password given a machine hostname and a challenge token.  To launch
the web service, run the following command:

```
potd serve
```

This will launch a webserver on port 1323 on all interfaces.  You can
influence the parameters of the server by setting the following
variables:

  * `LOG_LEVEL` - Log level to enable.  Supports `debug`, `info`,
    `warn`, `error`.  Defaults to `info`.
  * `POTD_BIND` - The address and port to bind to.  Defaults to
    `:1323`.
  * `POTD_DEBUG` - Enables template debugging.  Set to the path of the
    debug template cache to enable.
  * `POTD_DB` - Database engine to use.  Supported values: `sqlite`.
    Defaults to `sqlite`.
  * `POTD_SQLITE_PATH` - Path to the sqlite database.  Defaults to
    `potd.db`.
  * `POTD_CLIENT_CA` - Path to the client CA certificate.  Defaults to
    `client-ca.pem`.
  * `POTD_TLS_CERT` - Path to the server certificate.  Defaults to
    `tls.pem`.
  * `POTD_TLS_KEY` - Path to the server certificate key.  Defaults to
    `tls.key`.
  * `POTD_TRIM_PREFIX` - A string to be trimmed from the start of a
    client TLS CN.  This can assist with matching hostnames to
    certificates where the two may not be a perfect match at all
    times.  Default unset.
  * `POTD_TRIM_SUFFIX` - A string to be trimmed from the end of a
    client TLS CN.  This can assist with matching hostnames to
    certificates where the two may not be a perfect match at all
    times.  Default unset.

The trim variables are particularly useful where you may have special
DNS names for addresses that are guaranteed to resolve to a particular
internal network.  Take for example the case where a machine's
hostname is `box.example.com` but it obtains a certificate for
`box-internal.example.com`.  The domain component is unchecked since
we are comparing the hostname and not the FQDN.  So `potd` will try to
compare `box` with `box-internal` which does not match, but by setting
`POTD_TRIM_SUFFIX=-internal` then the comparison will succeed as `box`
= `box`.

Additionally, `potd serve` understands all environment variables that
can be set to influence
[AuthWare](https://github.com/the-maldridge/authware) to configure
authentication sources and parameters.
