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

142c7c67-9c46-4197-9f18-ddf100699ff1


examplebox login:
```

The UUID is the lynchpin to decoding the machine's root password
externally.  A user copies this token and the machine's hostname, and
uses the resolve function to obtain the password:

```
$ potd resolve --shared-token token.txt examplebox 142c7c67-9c46-4197-9f18-ddf100699ff1
4836C30DEEE73304352450279C8345291CACFDBB75A3C09D
```

The resolver combines the hostname, challenge token, and shared token
to generate the same random initialization vector and generate the
same password again.  The user can then use the revealed password to
log in.
