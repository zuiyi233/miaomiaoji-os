# RuleBack API 文档

> 本文档记录所有API接口，每次接口变更后需要更新此文档

---

## 通用说明

### 基础URL

```
开发环境: http://localhost:8080/api/v1
生产环境: https://api.example.com/api/v1
```

### 响应格式

所有接口返回统一的JSON格式：

```json
{
    "code": 0,
    "message": "success",
    "data": {}
}
```

### 状态码说明

| code | 说明 |
|------|------|
| 0 | 成功 |
| 10001 | 参数错误 |
| 10002 | 未认证 |
| 10003 | 无权限 |
| 10004 | 资源不存在 |
| 10005 | 资源冲突 |
| 10006 | 内部错误 |
| 10007 | 数据库错误 |
| 20001 | 用户不存在 |
| 20002 | 用户已存在 |
| 20003 | 用户已禁用 |
| 20004 | 密码错误 |
| 20005 | Token已过期 |
| 20006 | Token无效 |

### 分页响应

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "list": [],
        "total": 100,
        "page": 1,
        "page_size": 10,
        "total_pages": 10
    }
}
```

### 认证方式

需要认证的接口在Header中携带Token：

```
Authorization: Bearer <token>
```

---

## 系统接口

### 健康检查

```
GET /health
```

**响应:**
```json
{
    "status": "healthy"
}
```

### 存活检查

```
GET /ping
```

**响应:**
```
pong
```

---

## 用户接口

### 用户注册

```
POST /api/v1/auth/register
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名，3-50字符 |
| password | string | 是 | 密码，6-50字符 |
| nickname | string | 否 | 昵称，最大50字符 |
| email | string | 否 | 邮箱 |
| phone | string | 否 | 手机号 |

**请求示例:**
```json
{
    "username": "testuser",
    "password": "123456",
    "nickname": "测试用户",
    "email": "test@example.com"
}
```

**响应示例:**
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "id": 1,
        "username": "testuser",
        "nickname": "测试用户",
        "email": "test@example.com",
        "phone": "",
        "avatar": "",
        "status": 1,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
    }
}
```

### 用户登录

```
POST /api/v1/auth/login
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名 |
| password | string | 是 | 密码 |

**响应示例:**
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "user": {
            "id": 1,
            "username": "testuser",
            "nickname": "测试用户",
            "email": "test@example.com",
            "status": 1
        }
    }
}
```

### 获取当前用户信息

```
GET /api/v1/user/profile
```

**需要认证:** 是

**响应示例:**
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "id": 1,
        "username": "testuser",
        "nickname": "测试用户",
        "email": "test@example.com",
        "phone": "",
        "avatar": "",
        "status": 1
    }
}
```

### 更新用户信息

```
PUT /api/v1/user/profile
```

**需要认证:** 是

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| nickname | string | 否 | 昵称 |
| email | string | 否 | 邮箱 |
| phone | string | 否 | 手机号 |
| avatar | string | 否 | 头像URL |

---

## 分类接口

### 获取分类列表

```
GET /api/v1/categories
```

**查询参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| parent_id | int | 否 | 父分类ID，不传则获取所有分类 |

**响应示例:**
```json
{
    "code": 0,
    "message": "success",
    "data": [
        {
            "id": 1,
            "name": "电子产品",
            "parent_id": 0,
            "sort": 0,
            "status": 1
        }
    ]
}
```

### 获取分类详情

```
GET /api/v1/categories/:id
```

### 创建分类

```
POST /api/v1/categories
```

**需要认证:** 是

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 分类名称 |
| parent_id | int | 否 | 父分类ID |
| sort | int | 否 | 排序值 |

### 更新分类

```
PUT /api/v1/categories/:id
```

**需要认证:** 是

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 否 | 分类名称 |
| sort | int | 否 | 排序值 |
| status | int | 否 | 状态 0-禁用 1-启用 |

### 删除分类

```
DELETE /api/v1/categories/:id
```

**需要认证:** 是

---

## 商品接口

### 获取商品列表

```
GET /api/v1/products
```

**查询参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| page_size | int | 否 | 每页数量，默认10 |
| category_id | int | 否 | 分类ID |
| keyword | string | 否 | 搜索关键词 |
| status | int | 否 | 状态 0-下架 1-上架 |

**响应示例:**
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "list": [
            {
                "id": 1,
                "name": "iPhone 15",
                "description": "最新款苹果手机",
                "category_id": 1,
                "price": "6999.00",
                "stock": 100,
                "images": "https://example.com/1.jpg",
                "status": 1,
                "category": {
                    "id": 1,
                    "name": "手机"
                }
            }
        ],
        "total": 50,
        "page": 1,
        "page_size": 10,
        "total_pages": 5
    }
}
```

### 获取商品详情

```
GET /api/v1/products/:id
```

### 创建商品

```
POST /api/v1/products
```

**需要认证:** 是

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 商品名称 |
| description | string | 否 | 商品描述 |
| category_id | int | 是 | 分类ID |
| price | decimal | 是 | 价格 |
| stock | int | 否 | 库存，默认0 |
| images | string | 否 | 图片URL |

### 更新商品

```
PUT /api/v1/products/:id
```

**需要认证:** 是

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 否 | 商品名称 |
| description | string | 否 | 商品描述 |
| category_id | int | 否 | 分类ID |
| price | decimal | 否 | 价格 |
| stock | int | 否 | 库存 |
| images | string | 否 | 图片URL |
| status | int | 否 | 状态 0-下架 1-上架 |

### 删除商品

```
DELETE /api/v1/products/:id
```

**需要认证:** 是

---

## 购物车接口

### 获取购物车列表

```
GET /api/v1/cart
```

**需要认证:** 是

**响应示例:**
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "items": [
            {
                "id": 1,
                "product_id": 1,
                "name": "iPhone 15",
                "price": "6999.00",
                "images": "https://example.com/1.jpg",
                "quantity": 2,
                "stock": 100
            }
        ],
        "total_price": "13998.00",
        "total_count": 2
    }
}
```

### 添加商品到购物车

```
POST /api/v1/cart
```

**需要认证:** 是

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| product_id | int | 是 | 商品ID |
| quantity | int | 是 | 数量，最小1 |

### 更新购物车数量

```
PUT /api/v1/cart/:id
```

**需要认证:** 是

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| quantity | int | 是 | 数量，最小1 |

### 删除购物车项

```
DELETE /api/v1/cart/:id
```

**需要认证:** 是

### 清空购物车

```
DELETE /api/v1/cart
```

**需要认证:** 是

---

## 订单接口

### 获取订单列表

```
GET /api/v1/orders
```

**需要认证:** 是

**查询参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| page_size | int | 否 | 每页数量，默认10 |
| status | int | 否 | 订单状态 |

**订单状态说明:**

| 状态值 | 说明 |
|--------|------|
| 0 | 待支付 |
| 1 | 已支付 |
| 2 | 已发货 |
| 3 | 已完成 |
| 4 | 已取消 |

**响应示例:**
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "list": [
            {
                "id": 1,
                "order_no": "20240101120000abc12345",
                "user_id": 1,
                "total_amount": "13998.00",
                "status": 0,
                "address": "北京市朝阳区xxx",
                "remark": "",
                "items": [
                    {
                        "id": 1,
                        "order_id": 1,
                        "product_id": 1,
                        "name": "iPhone 15",
                        "price": "6999.00",
                        "quantity": 2
                    }
                ],
                "created_at": "2024-01-01T12:00:00Z"
            }
        ],
        "total": 10,
        "page": 1,
        "page_size": 10,
        "total_pages": 1
    }
}
```

### 创建订单

```
POST /api/v1/orders
```

**需要认证:** 是

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| address | string | 是 | 收货地址 |
| remark | string | 否 | 备注 |

**说明:** 从购物车中的商品创建订单，创建成功后会清空购物车

### 获取订单详情

```
GET /api/v1/orders/:id
```

**需要认证:** 是

### 取消订单

```
POST /api/v1/orders/:id/cancel
```

**需要认证:** 是

**说明:** 只能取消待支付状态的订单，取消后会恢复商品库存

### 支付订单

```
POST /api/v1/orders/:id/pay
```

**需要认证:** 是

**说明:** 模拟支付，将订单状态改为已支付

### 确认收货

```
POST /api/v1/orders/:id/confirm
```

**需要认证:** 是

**说明:** 用户确认收货，只能确认已发货状态的订单

---

## 收货地址接口

### 获取地址列表

```
GET /api/v1/addresses
```

**需要认证:** 是

**响应示例:**
```json
{
    "code": 0,
    "message": "success",
    "data": [
        {
            "id": 1,
            "user_id": 1,
            "name": "张三",
            "phone": "13800138000",
            "province": "北京市",
            "city": "北京市",
            "district": "朝阳区",
            "detail": "xx街道xx号",
            "is_default": true
        }
    ]
}
```

### 创建地址

```
POST /api/v1/addresses
```

**需要认证:** 是

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 收货人姓名 |
| phone | string | 是 | 联系电话 |
| province | string | 是 | 省份 |
| city | string | 是 | 城市 |
| district | string | 是 | 区县 |
| detail | string | 是 | 详细地址 |
| is_default | bool | 否 | 是否默认地址 |

### 获取地址详情

```
GET /api/v1/addresses/:id
```

**需要认证:** 是

### 获取默认地址

```
GET /api/v1/addresses/default
```

**需要认证:** 是

### 更新地址

```
PUT /api/v1/addresses/:id
```

**需要认证:** 是

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 否 | 收货人姓名 |
| phone | string | 否 | 联系电话 |
| province | string | 否 | 省份 |
| city | string | 否 | 城市 |
| district | string | 否 | 区县 |
| detail | string | 否 | 详细地址 |
| is_default | bool | 否 | 是否默认地址 |

### 删除地址

```
DELETE /api/v1/addresses/:id
```

**需要认证:** 是

### 设置默认地址

```
POST /api/v1/addresses/:id/default
```

**需要认证:** 是

---

## 商品收藏接口

### 获取收藏列表

```
GET /api/v1/favorites
```

**需要认证:** 是

**查询参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| page_size | int | 否 | 每页数量，默认10 |

**响应示例:**
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "list": [
            {
                "id": 1,
                "user_id": 1,
                "product_id": 1,
                "created_at": "2024-01-01T12:00:00Z",
                "product": {
                    "id": 1,
                    "name": "iPhone 15",
                    "price": "6999.00",
                    "images": "https://example.com/1.jpg",
                    "status": 1
                }
            }
        ],
        "total": 5,
        "page": 1,
        "page_size": 10,
        "total_pages": 1
    }
}
```

### 添加收藏

```
POST /api/v1/favorites
```

**需要认证:** 是

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| product_id | int | 是 | 商品ID |

### 取消收藏

```
DELETE /api/v1/favorites/:product_id
```

**需要认证:** 是

### 检查是否已收藏

```
GET /api/v1/favorites/:product_id/check
```

**需要认证:** 是

**响应示例:**
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "is_favorited": true
    }
}
```

---

## 商品评价接口

### 创建评价

```
POST /api/v1/reviews
```

**需要认证:** 是

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| order_id | int | 是 | 订单ID |
| product_id | int | 是 | 商品ID |
| rating | int | 是 | 评分，1-5 |
| content | string | 是 | 评价内容 |
| images | string | 否 | 评价图片URL，多个用逗号分隔 |

**说明:** 只能评价已完成的订单中的商品

### 获取我的评价列表

```
GET /api/v1/reviews/user
```

**需要认证:** 是

**查询参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| page_size | int | 否 | 每页数量，默认10 |

### 获取商品评价列表

```
GET /api/v1/products/:id/reviews
```

**需要认证:** 否

**查询参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| page_size | int | 否 | 每页数量，默认10 |

**响应示例:**
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "list": [
            {
                "id": 1,
                "user_id": 1,
                "product_id": 1,
                "order_id": 1,
                "rating": 5,
                "content": "非常好的商品",
                "images": "",
                "created_at": "2024-01-01T12:00:00Z"
            }
        ],
        "total": 10,
        "page": 1,
        "page_size": 10,
        "total_pages": 1
    }
}
```

### 获取商品评分统计

```
GET /api/v1/products/:id/rating
```

**需要认证:** 否

**响应示例:**
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "avg_rating": 4.5,
        "review_count": 100
    }
}
```

---

## 评论接口

### 评论类型说明

| 类型值 | 说明 |
|--------|------|
| 1 | 商品评论 |
| 2 | 评价回复 |

### 获取评论列表

```
GET /api/v1/comments
```

**需要认证:** 否

**查询参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | int | 是 | 评论类型 1-商品评论 2-评价回复 |
| target_id | int | 是 | 目标ID（商品ID或评价ID） |
| page | int | 否 | 页码，默认1 |
| page_size | int | 否 | 每页数量，默认10 |

**响应示例:**
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "list": [
            {
                "id": 1,
                "user_id": 1,
                "type": 1,
                "target_id": 1,
                "parent_id": 0,
                "reply_uid": 0,
                "content": "这个商品很好",
                "created_at": "2024-01-01T12:00:00Z",
                "user": {
                    "id": 1,
                    "nickname": "用户A"
                },
                "replies": [
                    {
                        "id": 2,
                        "user_id": 2,
                        "content": "同意",
                        "user": {"id": 2, "nickname": "用户B"}
                    }
                ]
            }
        ],
        "total": 10,
        "page": 1,
        "page_size": 10,
        "total_pages": 1
    }
}
```

### 创建评论

```
POST /api/v1/comments
```

**需要认证:** 是

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | int | 是 | 评论类型 1-商品评论 2-评价回复 |
| target_id | int | 是 | 目标ID |
| parent_id | int | 否 | 父评论ID（回复时使用） |
| reply_uid | int | 否 | 被回复用户ID |
| content | string | 是 | 评论内容，最大500字符 |

### 获取评论回复列表

```
GET /api/v1/comments/:id/replies
```

**需要认证:** 否

**查询参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| page_size | int | 否 | 每页数量，默认10 |

### 删除评论

```
DELETE /api/v1/comments/:id
```

**需要认证:** 是

**说明:** 只能删除自己的评论，删除时会同时删除所有子回复

### 获取评论数量

```
GET /api/v1/comments/count
```

**需要认证:** 否

**查询参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | int | 是 | 评论类型 |
| target_id | int | 是 | 目标ID |

**响应示例:**
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "count": 100
    }
}
```

---

## 管理员接口

> 管理员接口需要用户角色为 `admin`

### 获取所有订单

```
GET /api/v1/admin/orders
```

**需要认证:** 是（管理员）

**查询参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| page_size | int | 否 | 每页数量，默认10 |
| status | int | 否 | 订单状态 |
| user_id | int | 否 | 用户ID |
| order_no | string | 否 | 订单号（模糊搜索） |

### 获取订单详情

```
GET /api/v1/admin/orders/:id
```

**需要认证:** 是（管理员）

### 订单发货

```
POST /api/v1/admin/orders/:id/ship
```

**需要认证:** 是（管理员）

**说明:** 只能发货已支付状态的订单

### 订单完成

```
POST /api/v1/admin/orders/:id/complete
```

**需要认证:** 是（管理员）

**说明:** 只能完成已发货状态的订单

### 更新商品库存

```
PUT /api/v1/admin/products/:id/stock
```

**需要认证:** 是（管理员）

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| stock | int | 是 | 库存数量 |

---

## 更新日志

| 日期 | 版本 | 变更内容 |
|------|------|---------|
| 2026-01-30 | v1.4 | 添加评论功能（商品评论、评价回复） |
| 2026-01-30 | v1.3 | 添加商品收藏、商品评价、用户确认收货功能 |
| 2026-01-30 | v1.2 | 添加收货地址管理、管理员接口、订单发货/完成功能 |
| 2026-01-30 | v1.1 | 添加商城模块：用户、分类、商品、购物车、订单 |
| 2026-01-30 | v1.0 | 框架初始版本 |
