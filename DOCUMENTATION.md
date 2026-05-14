# Integration Guide for Sub-Apps

This guide explains how to integrate your applications (`*.resultspro.ng`) with the centralized authentication service at `auth.resultspro.ng`.

## Base URL
The authentication service is available at `https://auth.resultspro.ng` (Internal port: `7000`).

## 1. Authentication Flows

### Google OAuth
1.  Redirect users to `https://auth.resultspro.ng/auth/google`.
2.  After successful login, the service will return a JSON response with `access_token` and `refresh_token`.
    *   *Note: In a standard web flow, you might want to handle the redirection back to your app by passing a `state` or `redirect_uri` (to be implemented if needed).*

### Microsoft OAuth
1.  Redirect users to `https://auth.resultspro.ng/auth/microsoft`.
2.  After successful login, the service will return a JSON response with `access_token` and `refresh_token`.

### Email/Password Signup
- **Endpoint**: `POST /auth/signup`
- **Body**:
  ```json
  {
    "email": "user@example.com",
    "password": "securepassword",
    "full_name": "John Doe"
  }
  ```

### Email/Password Login
- **Endpoint**: `POST /auth/login`
- **Body**:
  ```json
  {
    "email": "user@example.com",
    "password": "securepassword"
  }
  ```

---

## 2. Session Management

### Refresh Token
- **Endpoint**: `POST /auth/refresh`
- **Body**:
  ```json
  {
    "refresh_token": "your_refresh_token"
  }
  ```
- **Response**: Returns a new `access_token`.

### Logout (Current Device)
- **Endpoint**: `POST /auth/logout`
- **Body**:
  ```json
  {
    "refresh_token": "your_refresh_token"
  }
  ```

### Logout (All Devices)
- **Endpoint**: `POST /auth/logout-all`
- **Headers**: `Authorization: Bearer <access_token>`

---

## 3. Token Verification (Introspection)

Sub-apps should verify the user's session by calling the introspection endpoint. This ensures the token is valid and the account is not suspended.

- **Endpoint**: `POST /auth/introspect`
- **Headers**:
  - `X-App-ID`: Your unique App ID.
  - `X-App-Secret`: Your App Secret Key.
- **Body**:
  ```json
  {
    "token": "user_access_token"
  }
  ```
- **Success Response (200 OK)**:
  ```json
  {
    "active": true,
    "user": {
      "id": "global-user-uuid",
      "email": "user@example.com",
      "full_name": "John Doe",
      "account_status": "active"
    }
  }
  ```
- **Failure Response**:
  ```json
  {
    "active": false,
    "reason": "account_suspended" (optional)
  }
  ```

---

## 4. Account Management

- **Verify Email**: `POST /auth/verify-email` with `{ "token": "..." }`
- **Forgot Password**: `POST /auth/forgot-password` with `{ "email": "..." }`
- **Reset Password**: `POST /auth/reset-password` with `{ "token": "...", "new_password": "..." }`
- **Change Password**: `POST /auth/change-password` (Auth required) with `{ "old_password": "...", "new_password": "..." }`
- **Change Email**: `POST /auth/change-email` (Auth required) with `{ "new_email": "..." }`
