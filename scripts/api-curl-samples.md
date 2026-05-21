# API curl samples

Base URL defaults to `http://127.0.0.1:8080`. Override with:

```bash
export BASE_URL=http://127.0.0.1:8080
export USER_ID=4a2bc1d8-7e3f-412e-a19b-625d91c84f32
```

Run the full happy-path flow:

```bash
chmod +x scripts/api-curl-samples.sh
./scripts/api-curl-samples.sh
```

Or run one endpoint at a time:

```bash
./scripts/api-curl-samples.sh health
./scripts/api-curl-samples.sh login
```

---

## Health

```bash
curl -sS "${BASE_URL}/health"
```

## Auth — login

```bash
curl -sS -X POST "${BASE_URL}/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"4a2bc1d8-7e3f-412e-a19b-625d91c84f32"}'
```

Save tokens:

```bash
export ACCESS_TOKEN="<access_token from response>"
export REFRESH_TOKEN="<refresh_token from response>"
```

## Auth — refresh

```bash
curl -sS -X POST "${BASE_URL}/api/v1/auth/refresh" \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"${REFRESH_TOKEN}\"}"
```

## Images — bulk upload (202 Accepted)

```bash
curl -sS -X POST "${BASE_URL}/api/v1/images/bulk" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "images": [
      { "original_filename": "landscape.png", "file_type": "image/png" },
      { "original_filename": "portrait.jpg", "file_type": "image/jpeg" }
    ]
  }'
```

Allowed `file_type` values: `image/png`, `image/jpeg`, `image/gif`, `image/webp`, `image/bmp`, `image/tiff`.

Save an image id from the response:

```bash
export IMAGE_ID="<image_id from records[0]>"
```

## Images — list

```bash
curl -sS "${BASE_URL}/api/v1/images" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

## Images — get by id

```bash
curl -sS "${BASE_URL}/api/v1/images/${IMAGE_ID}" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

## Images — download URL

```bash
curl -sS "${BASE_URL}/api/v1/images/${IMAGE_ID}/download" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

## Images — update metadata

```bash
curl -sS -X PUT "${BASE_URL}/api/v1/images/${IMAGE_ID}" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"original_filename":"updated_v2_landscape.png"}'
```

## Images — delete (204 No Content)

```bash
curl -sS -w "\nHTTP %{http_code}\n" -X DELETE \
  "${BASE_URL}/api/v1/images/${IMAGE_ID}" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```
