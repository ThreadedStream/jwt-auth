# jwt-auth

Do not forget to export secret key beforehand.

Linux:
```bash
  export SECRET_KEY=<your_secret_key>
```
Windows:
```batch
  SET SECRET_KEY=<your_secret_key>
```

There's 2 endpoints in total, namely /signin (to get a pair of access and refresh tokens), and /refresh (to refresh access token)

Server runs on localhost with port 4560
