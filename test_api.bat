@echo off
set BASE_URL=http://localhost:8080/api/v1

echo ========================================
echo Coupon Service API Test (Windows)
echo ========================================

echo.
echo [1] Create new user template ^(10 yuan off, no threshold^)
curl -s -X POST "%BASE_URL%/templates" ^
  -H "Content-Type: application/json" ^
  -d "{\"name\":\"新人礼包\",\"type\":1,\"value\":10,\"threshold\":0,\"total_count\":1000,\"per_user_limit\":1,\"valid_from\":\"2024-01-01T00:00:00Z\",\"valid_to\":\"2026-12-31T23:59:59Z\",\"applicable_level\":0,\"category\":0}"
echo.

echo.
echo [2] Create general template ^(50 off 200^)
curl -s -X POST "%BASE_URL%/templates" ^
  -H "Content-Type: application/json" ^
  -d "{\"name\":\"满200减50\",\"type\":0,\"value\":50,\"threshold\":200,\"total_count\":100,\"per_user_limit\":1,\"valid_from\":\"2024-01-01T00:00:00Z\",\"valid_to\":\"2026-12-31T23:59:59Z\",\"applicable_level\":0,\"category\":0}"
echo.

echo.
echo [3] List templates
curl -s "%BASE_URL%/templates?page=1^&size=10"
echo.

echo.
echo [4] New user register ^(user 1001^) - should get new user gift
curl -s -X POST "%BASE_URL%/users/register" ^
  -H "Content-Type: application/json" ^
  -d "{\"user_id\":1001,\"level\":0}"
echo.

echo.
echo [5] Claim coupon template 2 for user 1001
curl -s -X POST "%BASE_URL%/coupons/claim" ^
  -H "Content-Type: application/json" ^
  -d "{\"template_id\":2,\"user_id\":1001}"
echo.

echo.
echo [6] List user coupons
curl -s "%BASE_URL%/users/1001/coupons"
echo.

echo.
echo [7] Use coupon ^(record_id 2, order_amount 300^)
curl -s -X POST "%BASE_URL%/coupons/use" ^
  -H "Content-Type: application/json" ^
  -d "{\"record_id\":2,\"user_id\":1001,\"order_amount\":300,\"order_id\":\"ORD20240001\"}"
echo.

echo.
echo [8] Try use again - should fail
curl -s -X POST "%BASE_URL%/coupons/use" ^
  -H "Content-Type: application/json" ^
  -d "{\"record_id\":2,\"user_id\":1001,\"order_amount\":300,\"order_id\":\"ORD20240002\"}"
echo.

echo.
echo [9] Try use with low amount - should return 412
curl -s -X POST "%BASE_URL%/coupons/use" ^
  -H "Content-Type: application/json" ^
  -d "{\"record_id\":1,\"user_id\":1001,\"order_amount\":100,\"order_id\":\"ORD20240003\"}"
echo.

echo.
echo ========================================
echo API Test Complete!
echo ========================================
pause
