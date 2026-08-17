#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:18095/api/v1}"
HEALTH_URL="${HEALTH_URL:-http://127.0.0.1:18095/healthz}"
ADMIN_USER="${SEED_ADMIN_USERNAME:-admin}"
ADMIN_PASS="${SEED_ADMIN_PASSWORD:-Admin123!}"
SUFFIX="$(date +%s)"

call() {
  local method="$1"
  local path="$2"
  local token="${3:-}"
  local data="${4:-}"
  local args=(-sS -w $'\n%{http_code}' -X "$method" "${BASE_URL}${path}" -H 'Content-Type: application/json')
  if [[ -n "$token" ]]; then
    args+=(-H "Authorization: Bearer ${token}")
  fi
  if [[ -n "$data" ]]; then
    args+=(-d "$data")
  fi
  curl "${args[@]}"
}

split_response() {
  local response="$1"
  STATUS="$(printf '%s\n' "$response" | tail -n1)"
  BODY="$(printf '%s\n' "$response" | sed '$d')"
}

expect_status() {
  local name="$1"
  local expected="$2"
  if [[ "$STATUS" != "$expected" ]]; then
    printf 'FAIL %-34s expected=%s actual=%s body=%s\n' "$name" "$expected" "$STATUS" "$BODY"
    exit 1
  fi
  printf 'PASS %-34s HTTP %s\n' "$name" "$STATUS"
}

expect_json() {
  local name="$1"
  shift
  if ! printf '%s' "$BODY" | jq -e "$@" >/dev/null; then
    printf 'FAIL %-34s filter=%s body=%s\n' "$name" "$*" "$BODY"
    exit 1
  fi
  printf 'PASS %-34s %s\n' "$name" "$(printf '%s' "$BODY" | jq -c "$@")"
}

response="$(curl -sS -w $'\n%{http_code}' "$HEALTH_URL")"
split_response "$response"
expect_status "health" "200"
expect_json "health_json" '.status == "ok"'

response="$(call POST /auth/login "" "{\"username\":\"${ADMIN_USER}\",\"password\":\"${ADMIN_PASS}\"}")"
split_response "$response"
expect_status "admin_login" "200"
ADMIN_TOKEN="$(printf '%s' "$BODY" | jq -r '.token')"
expect_json "admin_login_role" '.user.role == "admin"'

USER_NAME="visitor_${SUFFIX}"
response="$(call POST /auth/register "" "{\"username\":\"${USER_NAME}\",\"password\":\"Visitor123!\"}")"
split_response "$response"
expect_status "user_register" "201"
expect_json "user_register_role" '.role == "user" and .created_at != "0001-01-01T00:00:00Z"'

response="$(call POST /auth/login "" "{\"username\":\"${USER_NAME}\",\"password\":\"Visitor123!\"}")"
split_response "$response"
expect_status "user_login" "200"
USER_TOKEN="$(printf '%s' "$BODY" | jq -r '.token')"
expect_json "user_login_role" '.user.role == "user"'

CAT_NAME="curl-分类-${SUFFIX}"
response="$(call POST /admin/categories "$ADMIN_TOKEN" "{\"name\":\"${CAT_NAME}\",\"description\":\"curl verification\"}")"
split_response "$response"
expect_status "category_create" "201"
CAT_ID="$(printf '%s' "$BODY" | jq -r '.id')"
expect_json "category_create_times" '.created_at != "0001-01-01T00:00:00Z"'

response="$(call PUT "/admin/categories/${CAT_ID}" "$ADMIN_TOKEN" "{\"name\":\"${CAT_NAME}-updated\",\"description\":\"updated\"}")"
split_response "$response"
expect_status "category_update" "200"

response="$(call GET /categories)"
split_response "$response"
expect_status "category_public_list" "200"
expect_json "category_public_list_contains" --argjson id "$CAT_ID" 'any(.[]; .id == $id)'

TAG_NAME="curl-标签-${SUFFIX}"
response="$(call POST /admin/tags "$ADMIN_TOKEN" "{\"name\":\"${TAG_NAME}\"}")"
split_response "$response"
expect_status "tag_create" "201"
TAG_ID="$(printf '%s' "$BODY" | jq -r '.id')"

response="$(call PUT "/admin/tags/${TAG_ID}" "$ADMIN_TOKEN" "{\"name\":\"${TAG_NAME}-updated\"}")"
split_response "$response"
expect_status "tag_update" "200"

response="$(call GET /tags)"
split_response "$response"
expect_status "tag_public_list" "200"
expect_json "tag_public_list_contains" --argjson id "$TAG_ID" 'any(.[]; .id == $id)'

KEYWORD="CURL-UNIQUE-${SUFFIX}"
ARTICLE_PAYLOAD="{\"category_id\":1,\"title\":\"curl 验证文章 ${SUFFIX}\",\"excerpt\":\"草稿摘要\",\"content\":\"这篇正文包含 ${KEYWORD}，用于完整验证。\",\"status\":\"draft\",\"tag_ids\":[1,2]}"
response="$(call POST /admin/articles "$ADMIN_TOKEN" "$ARTICLE_PAYLOAD")"
split_response "$response"
expect_status "article_draft_create" "201"
ARTICLE_ID="$(printf '%s' "$BODY" | jq -r '.id')"
expect_json "article_draft_status" '.status == "draft"'

response="$(call GET "/articles?q=${KEYWORD}")"
split_response "$response"
expect_status "draft_not_searchable" "200"
expect_json "draft_search_empty" 'length == 0'

response="$(call GET "/articles/${ARTICLE_ID}")"
split_response "$response"
expect_status "draft_public_get_hidden" "404"

response="$(call GET "/admin/articles?status=draft" "$ADMIN_TOKEN")"
split_response "$response"
expect_status "admin_draft_list" "200"
expect_json "admin_draft_list_contains" --argjson id "$ARTICLE_ID" 'any(.[]; .id == $id)'

response="$(call PATCH "/admin/articles/${ARTICLE_ID}/status" "$ADMIN_TOKEN" '{"status":"published"}')"
split_response "$response"
expect_status "article_publish" "200"
expect_json "article_published_status" '.status == "published" and .published_at != null'

response="$(call GET "/articles/${ARTICLE_ID}")"
split_response "$response"
expect_status "published_public_get" "200"
expect_json "published_public_get_content" --arg keyword "$KEYWORD" '.content | contains($keyword)'

response="$(call GET "/articles?category_id=1")"
split_response "$response"
expect_status "browse_by_category" "200"
expect_json "browse_category_contains" --argjson id "$ARTICLE_ID" 'any(.[]; .id == $id)'

response="$(call GET "/articles?tag_id=2")"
split_response "$response"
expect_status "browse_by_tag" "200"
expect_json "browse_tag_contains" --argjson id "$ARTICLE_ID" 'any(.[]; .id == $id)'

response="$(call GET "/articles?q=${KEYWORD}")"
split_response "$response"
expect_status "search_published" "200"
expect_json "search_published_contains" --argjson id "$ARTICLE_ID" 'any(.[]; .id == $id)'

UPDATE_PAYLOAD="{\"category_id\":2,\"title\":\"curl 验证文章 ${SUFFIX}-updated\",\"slug\":\"\",\"excerpt\":\"更新摘要\",\"content\":\"更新后的正文，仍包含 ${KEYWORD}。\",\"status\":\"published\",\"tag_ids\":[3]}"
response="$(call PUT "/admin/articles/${ARTICLE_ID}" "$ADMIN_TOKEN" "$UPDATE_PAYLOAD")"
split_response "$response"
expect_status "article_update" "200"
expect_json "article_update_status" '.status == "published"'

response="$(call POST "/articles/${ARTICLE_ID}/comments" "" '{"author_name":"访客甲","author_email":"a@example.com","content":"待审核评论"}')"
split_response "$response"
expect_status "comment_create_pending" "201"
COMMENT1_ID="$(printf '%s' "$BODY" | jq -r '.id')"
expect_json "comment_create_pending_status" '.status == "pending"'

response="$(call GET "/articles/${ARTICLE_ID}/comments")"
split_response "$response"
expect_status "comments_hidden_before_approve" "200"
expect_json "comments_empty_before_approve" 'length == 0'

response="$(call GET "/admin/comments?status=pending" "$ADMIN_TOKEN")"
split_response "$response"
expect_status "admin_pending_list" "200"
expect_json "admin_pending_contains" --argjson id "$COMMENT1_ID" 'any(.[]; .id == $id)'

response="$(call PATCH "/admin/comments/${COMMENT1_ID}" "$ADMIN_TOKEN" '{"status":"approved"}')"
split_response "$response"
expect_status "comment_approve" "200"
expect_json "comment_approve_status" '.status == "approved"'

response="$(call GET "/articles/${ARTICLE_ID}/comments")"
split_response "$response"
expect_status "approved_comment_public" "200"
expect_json "approved_comment_public_contains" --argjson id "$COMMENT1_ID" 'any(.[]; .id == $id)'

response="$(call POST "/articles/${ARTICLE_ID}/comments" "" '{"author_name":"访客乙","author_email":"b@example.com","content":"将被拒绝的评论"}')"
split_response "$response"
expect_status "comment_create_for_reject" "201"
COMMENT2_ID="$(printf '%s' "$BODY" | jq -r '.id')"

response="$(call PATCH "/admin/comments/${COMMENT2_ID}" "$ADMIN_TOKEN" '{"status":"rejected"}')"
split_response "$response"
expect_status "comment_reject" "200"
expect_json "comment_reject_status" '.status == "rejected"'

response="$(call GET "/articles/${ARTICLE_ID}/comments")"
split_response "$response"
expect_status "rejected_comment_hidden" "200"
expect_json "rejected_comment_not_public" --argjson id "$COMMENT2_ID" 'all(.[]; .id != $id)'

response="$(call POST "/articles/${ARTICLE_ID}/comments" "" '{"author_name":"访客丙","content":"将被删除的评论"}')"
split_response "$response"
expect_status "comment_create_for_delete" "201"
COMMENT3_ID="$(printf '%s' "$BODY" | jq -r '.id')"

response="$(call DELETE "/admin/comments/${COMMENT3_ID}" "$ADMIN_TOKEN")"
split_response "$response"
expect_status "comment_delete" "204"

response="$(call GET "/admin/comments?status=pending" "$ADMIN_TOKEN")"
split_response "$response"
expect_status "pending_empty_after_delete" "200"
expect_json "pending_does_not_contain_deleted" --argjson id "$COMMENT3_ID" 'all(.[]; .id != $id)'

response="$(call POST "/articles/2/comments" "" '{"author_name":"访客丁","content":"草稿不应允许评论"}')"
split_response "$response"
expect_status "draft_comment_blocked" "404"

response="$(call POST /admin/categories "" '{"name":"unauthorized"}')"
split_response "$response"
expect_status "admin_endpoint_requires_token" "401"

response="$(call POST /admin/categories "$USER_TOKEN" '{"name":"unauthorized"}')"
split_response "$response"
expect_status "normal_user_forbidden" "403"

response="$(call GET /auth/me "$ADMIN_TOKEN")"
split_response "$response"
expect_status "admin_me" "200"
expect_json "admin_me_role" '.role == "admin"'

response="$(call GET /auth/me "$USER_TOKEN")"
split_response "$response"
expect_status "user_me" "200"
expect_json "user_me_role" '.role == "user"'

response="$(call DELETE "/admin/articles/${ARTICLE_ID}" "$ADMIN_TOKEN")"
split_response "$response"
expect_status "article_delete" "204"

response="$(call GET "/articles/${ARTICLE_ID}")"
split_response "$response"
expect_status "deleted_article_hidden" "404"

response="$(call DELETE "/admin/categories/${CAT_ID}" "$ADMIN_TOKEN")"
split_response "$response"
expect_status "category_delete" "204"

response="$(call DELETE "/admin/tags/${TAG_ID}" "$ADMIN_TOKEN")"
split_response "$response"
expect_status "tag_delete" "204"

response="$(call GET /categories)"
split_response "$response"
expect_status "category_list_after_delete" "200"
expect_json "deleted_category_absent" --argjson id "$CAT_ID" 'all(.[]; .id != $id)'

response="$(call GET /tags)"
split_response "$response"
expect_status "tag_list_after_delete" "200"
expect_json "deleted_tag_absent" --argjson id "$TAG_ID" 'all(.[]; .id != $id)'

printf '\nALL CHECKS PASSED\n'
