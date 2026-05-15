#!/usr/bin/env bash
set -euo pipefail

API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
APP_USERNAME="${APP_USERNAME:-admin}"
APP_PASSWORD="${APP_PASSWORD:-pw123}"

if ! command -v curl >/dev/null 2>&1; then
  echo "error: curl is required" >&2
  exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
  echo "error: jq is required" >&2
  exit 1
fi

ACCESS_TOKEN=""
REFRESH_TOKEN=""
RESPONSE_STATUS=""
RESPONSE_BODY=""
CUSTOM_TYPE_ID="api_test_type_$(date +%s)"
CUSTOM_ITEM_ID=""

log_step() {
  echo ""
  echo "==> $1"
}

do_request() {
  local method="$1"
  local path="$2"
  local body="${3:-}"
  local use_auth="${4:-true}"

  local headers=()
  headers+=("-H" "Accept: application/json")

  if [[ "${use_auth}" == "true" ]]; then
    headers+=("-H" "Authorization: Bearer ${ACCESS_TOKEN}")
  fi

  if [[ -n "${body}" ]]; then
    headers+=("-H" "Content-Type: application/json")
  fi

  local output
  if [[ -n "${body}" ]]; then
    output="$(curl -sS -X "${method}" "${API_BASE_URL}${path}" "${headers[@]}" --data "${body}" -w $'\n%{http_code}')"
  else
    output="$(curl -sS -X "${method}" "${API_BASE_URL}${path}" "${headers[@]}" -w $'\n%{http_code}')"
  fi

  RESPONSE_STATUS="${output##*$'\n'}"
  RESPONSE_BODY="${output%$'\n'*}"
}

assert_status() {
  local expected="$1"
  local context="$2"

  if [[ "${RESPONSE_STATUS}" != "${expected}" ]]; then
    echo "error: ${context} expected status ${expected}, got ${RESPONSE_STATUS}" >&2
    echo "response body: ${RESPONSE_BODY}" >&2
    exit 1
  fi
}

extract_json() {
  local jq_expr="$1"
  jq -er "${jq_expr}" <<<"${RESPONSE_BODY}"
}

assert_json_expr() {
  local jq_expr="$1"
  local context="$2"

  if ! jq -e "${jq_expr}" >/dev/null <<<"${RESPONSE_BODY}"; then
    echo "error: ${context} response failed jq assertion: ${jq_expr}" >&2
    echo "response body: ${RESPONSE_BODY}" >&2
    exit 1
  fi
}

# 1) Health
log_step "health check"
do_request "GET" "/health" "" "false"
assert_status "200" "health check"

# 2) Auth login
log_step "auth login"
do_request "POST" "/api/v1/auth/login" "{\"username\":\"${APP_USERNAME}\",\"password\":\"${APP_PASSWORD}\"}" "false"
assert_status "200" "auth login"
ACCESS_TOKEN="$(extract_json '.access_token')"
REFRESH_TOKEN="$(extract_json '.refresh_token')"

# 3) Auth refresh
log_step "auth refresh"
do_request "POST" "/api/v1/auth/refresh" "{\"refresh_token\":\"${REFRESH_TOKEN}\"}" "false"
assert_status "200" "auth refresh"
ACCESS_TOKEN="$(extract_json '.access_token')"
REFRESH_TOKEN="$(extract_json '.refresh_token')"

# 4) Settings
log_step "settings get"
do_request "GET" "/api/v1/settings"
assert_status "200" "get settings"

log_step "settings update"
do_request "PATCH" "/api/v1/settings" '{"weight_unit":"g","distance_unit":"km","temperature_unit":"c","volume_unit":"ml"}'
assert_status "200" "update settings"

# 5) Persons
log_step "persons list"
do_request "GET" "/api/v1/persons"
assert_status "200" "list persons"

log_step "person create"
do_request "POST" "/api/v1/persons" '{"name":"API Test Person","gender":"other","birthdate":"1990-01-01","body_weight_grams":70000,"conditioning_level":"average"}'
assert_status "201" "create person"
assert_json_expr '.conditioning_level == "average"' "create person"
assert_json_expr '.body_weight_grams == 70000' "create person"
PERSON_ID="$(extract_json '.id')"

log_step "person get"
do_request "GET" "/api/v1/persons/${PERSON_ID}"
assert_status "200" "get person"
assert_json_expr '.conditioning_level == "average"' "get person"

log_step "person update"
do_request "PATCH" "/api/v1/persons/${PERSON_ID}" '{"name":"API Test Person Updated","conditioning_level":"athletic","body_weight_grams":72000}'
assert_status "200" "update person"
assert_json_expr '.name == "API Test Person Updated"' "update person"
assert_json_expr '.conditioning_level == "athletic"' "update person"
assert_json_expr '.body_weight_grams == 72000' "update person"

# 6) Manufacturers
log_step "manufacturers list"
do_request "GET" "/api/v1/manufacturers"
assert_status "200" "list manufacturers"

log_step "manufacturer create"
do_request "POST" "/api/v1/manufacturers" '{"name":"API Test Manufacturer","website":"https://example.com"}'
assert_status "201" "create manufacturer"
MANUFACTURER_ID="$(extract_json '.id')"

log_step "manufacturer get"
do_request "GET" "/api/v1/manufacturers/${MANUFACTURER_ID}"
assert_status "200" "get manufacturer"

log_step "manufacturer update"
do_request "PATCH" "/api/v1/manufacturers/${MANUFACTURER_ID}" '{"name":"API Test Manufacturer Updated"}'
assert_status "200" "update manufacturer"

# 7) Items
log_step "items list"
do_request "GET" "/api/v1/items"
assert_status "200" "list items"

log_step "item create"
do_request "POST" "/api/v1/items" "{\"name\":\"API Test Item\",\"type\":\"consumable\",\"is_active\":true,\"manufacturer_id\":\"${MANUFACTURER_ID}\",\"weight_grams\":100,\"volume_ml\":50}"
assert_status "201" "create item"
assert_json_expr '.type == "consumable"' "create item"
assert_json_expr '.name == "API Test Item"' "create item"
assert_json_expr '.weight_grams == 100' "create item"
assert_json_expr '.volume_ml == 50' "create item"
ITEM_ID="$(extract_json '.id')"

log_step "item get"
do_request "GET" "/api/v1/items/${ITEM_ID}"
assert_status "200" "get item"
assert_json_expr '.type == "consumable"' "get item"
assert_json_expr '.name == "API Test Item"' "get item"

log_step "item update"
do_request "PATCH" "/api/v1/items/${ITEM_ID}" '{"name":"API Test Item Updated","weight_grams":110}'
assert_status "200" "update item"
assert_json_expr '.type == "consumable"' "update item"
assert_json_expr '.name == "API Test Item Updated"' "update item"
assert_json_expr '.weight_grams == 110' "update item"

log_step "items list after update"
do_request "GET" "/api/v1/items"
assert_status "200" "list items after update"
assert_json_expr 'any(.[]; .id == "'"${ITEM_ID}"'" and .type == "consumable" and .name == "API Test Item Updated")' "list items after update"

# 8) Item Types + Custom Item Attributes
log_step "item-types list"
do_request "GET" "/api/v1/item-types"
assert_status "200" "list item types"
assert_json_expr 'any(.[]; .id == "consumable")' "list item types"

log_step "item-type create"
do_request "POST" "/api/v1/item-types" "{\"id\":\"${CUSTOM_TYPE_ID}\",\"name\":\"API Test Custom Type\",\"description\":\"created by curl test\",\"base_profile\":\"electronics\"}"
assert_status "201" "create item type"
assert_json_expr '.id == "'"${CUSTOM_TYPE_ID}"'"' "create item type"
assert_json_expr '.name == "API Test Custom Type"' "create item type"
assert_json_expr '.is_system == false' "create item type"

log_step "item-type get"
do_request "GET" "/api/v1/item-types/${CUSTOM_TYPE_ID}"
assert_status "200" "get item type"
assert_json_expr '.id == "'"${CUSTOM_TYPE_ID}"'"' "get item type"

log_step "item-type update"
do_request "PATCH" "/api/v1/item-types/${CUSTOM_TYPE_ID}" '{"name":"API Test Custom Type Updated","description":"updated by curl test"}'
assert_status "200" "update item type"
assert_json_expr '.name == "API Test Custom Type Updated"' "update item type"

log_step "item-type fields replace"
do_request "PUT" "/api/v1/item-types/${CUSTOM_TYPE_ID}/fields" '{"fields":[{"field_key":"output_watts","field_label":"Output watts","data_type":"number","is_required":false,"sort_order":1,"unit":"W"},{"field_key":"has_usb_pd","field_label":"Has USB PD","data_type":"boolean","is_required":false,"sort_order":2},{"field_key":"battery_chemistry","field_label":"Battery chemistry","data_type":"enum","is_required":false,"sort_order":3,"enum_options":["li-ion","lifepo4"]}]}'
assert_status "200" "replace item type fields"
assert_json_expr 'length == 3' "replace item type fields"
assert_json_expr 'any(.[]; .field_key == "output_watts" and .data_type == "number")' "replace item type fields"
assert_json_expr 'any(.[]; .field_key == "has_usb_pd" and .data_type == "boolean")' "replace item type fields"
assert_json_expr 'any(.[]; .field_key == "battery_chemistry" and .data_type == "enum")' "replace item type fields"

log_step "item-type fields list"
do_request "GET" "/api/v1/item-types/${CUSTOM_TYPE_ID}/fields"
assert_status "200" "list item type fields"
assert_json_expr 'any(.[]; .field_key == "output_watts")' "list item type fields"

log_step "custom item create"
do_request "POST" "/api/v1/items" "{\"name\":\"API Test Custom Item\",\"type\":\"${CUSTOM_TYPE_ID}\",\"is_active\":true,\"manufacturer_id\":\"${MANUFACTURER_ID}\",\"weight_grams\":220,\"volume_ml\":330,\"attributes\":{\"output_watts\":30,\"has_usb_pd\":true,\"battery_chemistry\":\"li-ion\"}}"
assert_status "201" "create custom item"
assert_json_expr '.type == "'"${CUSTOM_TYPE_ID}"'"' "create custom item"
assert_json_expr '.attributes.output_watts == 30' "create custom item"
assert_json_expr '.attributes.has_usb_pd == true' "create custom item"
CUSTOM_ITEM_ID="$(extract_json '.id')"

log_step "custom item get"
do_request "GET" "/api/v1/items/${CUSTOM_ITEM_ID}"
assert_status "200" "get custom item"
assert_json_expr '.type == "'"${CUSTOM_TYPE_ID}"'"' "get custom item"
assert_json_expr '.attributes.battery_chemistry == "li-ion"' "get custom item"

log_step "custom item update"
do_request "PATCH" "/api/v1/items/${CUSTOM_ITEM_ID}" '{"name":"API Test Custom Item Updated","attributes":{"output_watts":45,"has_usb_pd":true,"battery_chemistry":"lifepo4"}}'
assert_status "200" "update custom item"
assert_json_expr '.name == "API Test Custom Item Updated"' "update custom item"
assert_json_expr '.attributes.output_watts == 45' "update custom item"
assert_json_expr '.attributes.battery_chemistry == "lifepo4"' "update custom item"

log_step "items list with custom item"
do_request "GET" "/api/v1/items"
assert_status "200" "list items with custom item"
assert_json_expr 'any(.[]; .id == "'"${CUSTOM_ITEM_ID}"'" and .type == "'"${CUSTOM_TYPE_ID}"'" and .attributes.output_watts == 45)' "list items with custom item"

# 9) Sets
log_step "sets list"
do_request "GET" "/api/v1/sets"
assert_status "200" "list sets"

log_step "set create"
do_request "POST" "/api/v1/sets" '{"name":"API Test Set","set_category":"consumable"}'
assert_status "201" "create set"
SET_ID="$(extract_json '.id')"

log_step "set get"
do_request "GET" "/api/v1/sets/${SET_ID}"
assert_status "200" "get set"

log_step "set update"
do_request "PATCH" "/api/v1/sets/${SET_ID}" '{"name":"API Test Set Updated","set_category":"consumable"}'
assert_status "200" "update set"

log_step "set-items list"
do_request "GET" "/api/v1/sets/${SET_ID}/items"
assert_status "200" "list set items"

log_step "set-item add"
do_request "POST" "/api/v1/sets/${SET_ID}/items" "{\"item_id\":\"${ITEM_ID}\",\"quantity\":2,\"notes\":\"test note\",\"sort_order\":1}"
assert_status "201" "add set item"

log_step "set-item update"
do_request "PATCH" "/api/v1/sets/${SET_ID}/items/${ITEM_ID}" '{"quantity":3,"notes":"updated note"}'
assert_status "200" "update set item"

# 10) Packs
log_step "packs list"
do_request "GET" "/api/v1/packs"
assert_status "200" "list packs"

log_step "pack create"
do_request "POST" "/api/v1/packs" "{\"name\":\"API Test Pack\",\"trip_type\":\"multi_day\",\"person_id\":\"${PERSON_ID}\"}"
assert_status "201" "create pack"
PACK_ID="$(extract_json '.id')"

log_step "pack get"
do_request "GET" "/api/v1/packs/${PACK_ID}"
assert_status "200" "get pack"

log_step "pack update"
do_request "PATCH" "/api/v1/packs/${PACK_ID}" '{"name":"API Test Pack Updated","trip_type":"overnight"}'
assert_status "200" "update pack"

log_step "pack-items list"
do_request "GET" "/api/v1/packs/${PACK_ID}/items"
assert_status "200" "list pack items"

log_step "pack-item add"
do_request "POST" "/api/v1/packs/${PACK_ID}/items" "{\"item_id\":\"${ITEM_ID}\",\"quantity\":1,\"carry_status\":\"packed\"}"
assert_status "201" "add pack item"

log_step "pack-item update"
do_request "PATCH" "/api/v1/packs/${PACK_ID}/items/${ITEM_ID}" '{"quantity":2,"carry_status":"worn"}'
assert_status "200" "update pack item"

log_step "pack-sets list"
do_request "GET" "/api/v1/packs/${PACK_ID}/sets"
assert_status "200" "list pack sets"

log_step "pack-set add"
do_request "POST" "/api/v1/packs/${PACK_ID}/sets" "{\"set_id\":\"${SET_ID}\"}"
assert_status "201" "add pack set"

log_step "pack-set remove"
do_request "DELETE" "/api/v1/packs/${PACK_ID}/sets/${SET_ID}"
assert_status "204" "remove pack set"

log_step "pack-item remove"
do_request "DELETE" "/api/v1/packs/${PACK_ID}/items/${ITEM_ID}"
assert_status "204" "remove pack item"

# 11) Start Fresh (danger zone)
log_step "start fresh wrong password"
do_request "POST" "/api/v1/settings/start-fresh" '{"password":"wrong-password"}'
assert_status "401" "start fresh wrong password"

log_step "start fresh confirmed"
do_request "POST" "/api/v1/settings/start-fresh" "{\"password\":\"${APP_PASSWORD}\"}"
assert_status "204" "start fresh confirmed"

log_step "verify persons reset"
do_request "GET" "/api/v1/persons"
assert_status "200" "list persons after start fresh"
assert_json_expr 'length == 0' "list persons after start fresh"

log_step "verify manufacturers reset"
do_request "GET" "/api/v1/manufacturers"
assert_status "200" "list manufacturers after start fresh"
assert_json_expr 'length == 0' "list manufacturers after start fresh"

log_step "verify items reset"
do_request "GET" "/api/v1/items"
assert_status "200" "list items after start fresh"
assert_json_expr 'length == 0' "list items after start fresh"

log_step "verify sets reset"
do_request "GET" "/api/v1/sets"
assert_status "200" "list sets after start fresh"
assert_json_expr 'length == 0' "list sets after start fresh"

log_step "verify packs reset"
do_request "GET" "/api/v1/packs"
assert_status "200" "list packs after start fresh"
assert_json_expr 'length == 0' "list packs after start fresh"

log_step "verify custom item type removed"
do_request "GET" "/api/v1/item-types/${CUSTOM_TYPE_ID}"
assert_status "404" "get custom item type after start fresh"

log_step "verify settings defaults restored"
do_request "GET" "/api/v1/settings"
assert_status "200" "get settings after start fresh"
assert_json_expr '.weight_unit == "g"' "settings defaults"
assert_json_expr '.distance_unit == "km"' "settings defaults"
assert_json_expr '.temperature_unit == "c"' "settings defaults"
assert_json_expr '.volume_unit == "ml"' "settings defaults"
assert_json_expr '.currency == "eur"' "settings defaults"

# 12) Auth logout
log_step "auth logout"
do_request "POST" "/api/v1/auth/logout" "" "false"
assert_status "204" "auth logout"

echo ""
echo "full API curl test completed successfully"
