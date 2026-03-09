python scripts/spec-format.py backend/internal/repo backend/internal/apitest backend/testutil

repo:
 ✔ stop repo create
 ✔ stop repo create with departed at
 ✔ stop repo get by id
 ✔ stop repo get by id wrong trip
 ✔ stop repo list by trip id
 ✔ stop repo list by trip id empty
 ✔ stop repo update
 ✔ stop repo update wrong trip
 ✔ stop repo delete
 ✔ stop repo delete wrong trip
 ✔ stop repo get by id includes tags
 ✔ stop repo get by id empty tags
 ✔ stop repo list by trip id paged includes tags
 ✔ tag repo upsert create
 ✔ tag repo upsert idempotent by slug
 ✔ tag repo list all
 ✔ tag repo list prefix
 ✔ tag repo list empty
 ✔ tag repo add to stop
 ✔ tag repo add to stop idempotent
 ✔ tag repo list by stop
 ✔ tag repo list by stop empty
 ✔ tag repo remove from stop
 ✔ tag repo remove from stop not found
 ✔ tag repo update name ok
 ✔ tag repo update name not found
 ✔ tag repo delete ok
 ✔ tag repo delete not found
 ✔ trip repo create
 ✔ trip repo create nil end date
 ✔ trip repo get by id
 ✔ trip repo get by id not found
 ✔ trip repo list
 ✔ trip repo update
 ✔ trip repo update not found
 ✔ trip repo delete
 ✔ trip repo delete not found

apitest:
 ✔ trip create
 ✔ trip get by id
 ✔ trip get by id not found
 ✔ trip list
 ✔ trip update
 ✔ trip delete

testutil:
 ✔ migration006 fix stop dates midnight to noon est
 ✔ migrations