# 云存储统一管理工具 PRD (Product Requirements Document)

## 目录

### 1. 项目概述
- 1.1 项目名称
- 1.2 项目愿景
- 1.3 核心价值

### 2. 功能需求
- 2.1 核心功能模块
  - 2.1.1 桶（Bucket）管理
  - 2.1.2 对象（Object）管理
  - 2.1.3 同步功能
- 2.2 命令设计
  - 2.2.1 基础命令结构
  - 2.2.2 具体命令示例
  - 2.2.3 桶管理操作详细参考
  - 2.2.4 对象操作详细参考
  - 2.2.5 全局选项
  - 2.2.6 环境变量配置

### 3. 非功能需求
- 3.1 性能要求
- 3.2 可靠性要求
- 3.3 安全性要求
- 3.4 易用性要求

### 4. 技术架构
- 4.1 技术栈
- 4.2 架构设计
- 4.3 统一接口设计

### 5. 配置管理
- 5.1 配置文件格式
- 5.2 配置来源优先级

### 6. 项目计划
- 6.1 开发阶段
- 6.2 里程碑

### 7. 成功标准
- 7.1 功能标准
- 7.2 性能标准
- 7.3 质量标准

### 8. 风险与缓解
- 8.1 技术风险
- 8.2 安全风险
- 8.3 运营风险

### 9. 扩展计划

### 下一步行动建议

---

## 1. 项目概述

### 1.1 项目名称
`cloud-storage-tool` - 多云存储统一管理工具

### 1.2 项目愿景
开发一个统一的命令行工具，同时支持腾讯云COS和AWS S3，实现跨云存储服务的数据管理，简化多云环境下的存储操作。

### 1.3 核心价值
- **统一接口**：一套命令管理多个云存储服务
- **简化操作**：避免记忆不同云服务的CLI语法
- **提高效率**：批量操作、同步功能提升工作效率
- **降低成本**：优化存储使用，避免厂商锁定

## 2. 功能需求

### 2.1 核心功能模块

#### 2.1.1 桶（Bucket）管理
- **桶列表**：列出所有桶
- **桶创建**：创建新桶
- **桶删除**：删除空桶或强制删除
- **桶信息**：查看桶详细信息（区域、创建时间等）
- **桶权限**：管理桶的访问权限
- **桶清单配置**：配置桶的清单规则，生成存储对象清单报告
- **生命周期规则管理**：配置对象的生命周期规则，自动转换存储类型或过期删除

#### 2.1.2 对象（Object）管理
- **对象上传**：支持文件、目录上传
  - 单文件上传
  - 目录递归上传
  - 分块上传（大文件）
  - 断点续传
- **对象下载**：
  - 单文件下载
  - 目录递归下载
  - 选择性下载（通配符匹配）
- **对象删除**：
  - 单文件删除
  - 批量删除
  - 按前缀删除
- **对象列表**：
  - 列出桶内所有对象
  - 按前缀过滤
  - 分页显示
- **对象信息**：
  - 查看对象元数据
  - 查看对象大小、最后修改时间
  - 生成预签名URL

#### 2.1.3 同步功能
- **桶间同步**：COS ↔ S3 双向同步
- **本地同步**：本地 ↔ 云存储 双向同步
- **增量同步**：仅同步变化的文件
- **排除模式**：支持.gitignore风格排除规则

### 2.2 命令设计

#### 2.2.1 基础命令结构
```bash
# 全局结构
cloud-storage-tool [global-flags] <command> [command-flags]

# 服务选择
cloud-storage-tool --provider cos|s3 <command>

# 或使用子命令
cloud-storage-tool cos <command>
cloud-storage-tool s3 <command>
```

#### 2.2.2 具体命令示例
```bash
# 配置管理
cloud-storage-tool config init
cloud-storage-tool config list
cloud-storage-tool config test

# 桶操作
cloud-storage-tool cos ls
cloud-storage-tool cos mb my-bucket --region ap-singapore
cloud-storage-tool cos rb my-bucket --force

# 对象操作
cloud-storage-tool s3 cp local/file.txt s3://my-bucket/path/
cloud-storage-tool cos cp cos://bucket/path/ local/dir/ --recursive
cloud-storage-tool s3 rm s3://bucket/path/* --recursive

# 同步操作
cloud-storage-tool sync cos://bucket1/path/ s3://bucket2/path/
cloud-storage-tool sync local/dir/ cos://bucket/path/ --exclude "*.tmp"

# 信息查询
cloud-storage-tool cos stat cos://bucket/file.txt
cloud-storage-tool s3 presign s3://bucket/file.txt --expires 3600
```

#### 2.2.3 桶管理操作详细参考

##### 1. 桶列表 (List Buckets)

**命令格式**
```bash
cloud-storage-tool <provider> ls [options]
```

**参数说明**
| 参数 | 缩写 | 说明 | 默认值 |
|------|------|------|--------|
| `--region` | `-r` | 指定区域（列出该区域的所有桶） | 默认区域 |
| `--format` | `-f` | 输出格式（json, table, csv） | table |
| `--limit` | `-l` | 最多显示的桶数量 | 无限制 |

**使用示例**
```bash
# 列出所有COS桶
cloud-storage-tool cos ls

# 列出新加坡区域的所有S3桶
cloud-storage-tool s3 ls --region ap-singapore

# 以JSON格式输出COS桶列表
cloud-storage-tool cos ls --format json

# 列出前10个桶
cloud-storage-tool cos ls --limit 10
```

##### 2. 桶创建 (Create Bucket)

**命令格式**
```bash
cloud-storage-tool <provider> mb <bucket-name> [options]
```

**参数说明**
| 参数 | 缩写 | 说明 | 默认值 |
|------|------|------|--------|
| `--region` | `-r` | 桶所在的区域 | 默认区域 |
| `--acl` | `-a` | 访问控制权限 | private |
| `--storage-class` | `-s` | 存储类型（COS: STANDARD/STANDARD_IA/ARCHIVE, S3: STANDARD/STANDARD_IA/ONEZONE_IA/GLACIER） | STANDARD |
| `--versioning` | `-v` | 是否启用版本控制 | false |
| `--dry-run` | `-d` | 模拟运行，不实际创建 | false |

**ACL权限选项**
- **COS**: `private`, `public-read`, `public-read-write`
- **S3**: `private`, `public-read`, `public-read-write`, `authenticated-read`

**使用示例**
```bash
# 在新加坡区域创建私有桶
cloud-storage-tool cos mb my-backup-bucket --region ap-singapore

# 创建支持版本控制的S3桶
cloud-storage-tool s3 mb logs-bucket --region us-east-1 --versioning true

# 创建公开只读桶（COS）
cloud-storage-tool cos mb public-data --region ap-shanghai --acl public-read

# 创建低频访问存储桶
cloud-storage-tool s3 mb archive-bucket --region us-west-2 --storage-class STANDARD_IA

# 模拟创建（不实际执行）
cloud-storage-tool cos mb test-bucket --dry-run
```

##### 3. 桶删除 (Delete Bucket)

**命令格式**
```bash
cloud-storage-tool <provider> rb <bucket-name> [options]
```

**参数说明**
| 参数 | 缩写 | 说明 | 默认值 |
|------|------|------|--------|
| `--force` | `-f` | 强制删除非空桶 | false |
| `--region` | `-r` | 桶所在的区域 | 自动检测 |
| `--dry-run` | `-d` | 模拟运行，不实际删除 | false |
| `--recursive` | `-R` | 递归删除桶内所有对象后删除桶 | false |

**使用示例**
```bash
# 删除空桶
cloud-storage-tool cos rb my-bucket

# 强制删除非空桶
cloud-storage-tool s3 rb logs-bucket --force

# 递归删除桶内所有对象后删除桶
cloud-storage-tool cos rb temp-bucket --recursive

# 模拟删除（不实际执行）
cloud-storage-tool s3 rb test-bucket --dry-run

# 删除指定区域的桶
cloud-storage-tool cos rb old-bucket --region ap-beijing
```

##### 4. 桶信息查看 (Bucket Information)

**命令格式**
```bash
cloud-storage-tool <provider> info <bucket-name> [options]
```

**参数说明**
| 参数 | 缩写 | 说明 | 默认值 |
|------|------|------|--------|
| `--region` | `-r` | 桶所在的区域 | 自动检测 |
| `--format` | `-f` | 输出格式（json, yaml, table） | table |
| `--details` | `-d` | 显示详细信息 | false |

**显示信息包括**
- 桶名称、区域、创建时间
- 存储类型、访问权限
- 版本控制状态
- 对象数量、存储空间使用量
- 生命周期规则数量
- 清单配置状态

**使用示例**
```bash
# 查看桶基本信息
cloud-storage-tool cos info my-bucket

# 查看桶详细信息
cloud-storage-tool s3 info logs-bucket --details

# 以JSON格式输出桶信息
cloud-storage-tool cos info data-bucket --format json

# 查看指定区域的桶信息
cloud-storage-tool s3 info archive-bucket --region eu-west-1
```

##### 5. 桶权限管理 (Bucket ACL)

**命令格式**
```bash
cloud-storage-tool <provider> acl <bucket-name> <action> [options]
```

**子命令**
| 子命令 | 说明 |
|--------|------|
| `get` | 获取桶的ACL设置 |
| `set` | 设置桶的ACL |
| `grant` | 授予特定用户/用户组权限 |
| `revoke` | 撤销特定用户/用户组权限 |

**通用参数**
| 参数 | 缩写 | 说明 | 默认值 |
|------|------|------|--------|
| `--region` | `-r` | 桶所在的区域 | 自动检测 |
| `--format` | `-f` | 输出格式（json, table） | table |

**set子命令参数**
| 参数 | 缩写 | 说明 | 必需 |
|------|------|------|------|
| `--acl` | `-a` | ACL权限字符串 | 是 |

**grant/revoke子命令参数**
| 参数 | 缩写 | 说明 | 必需 |
|------|------|------|------|
| `--grantee` | `-g` | 被授权者（邮箱、用户ID等） | 是 |
| `--permission` | `-p` | 权限类型（READ, WRITE, FULL_CONTROL等） | 是 |

**使用示例**
```bash
# 获取桶的ACL设置
cloud-storage-tool cos acl my-bucket get

# 设置桶为公开只读
cloud-storage-tool s3 acl public-bucket set --acl public-read

# 授予特定用户读取权限（S3）
cloud-storage-tool s3 acl shared-bucket grant \
  --grantee user@example.com \
  --permission READ

# 撤销用户权限
cloud-storage-tool cos acl project-bucket revoke \
  --grantee 123456789012 \
  --permission WRITE

# 以JSON格式查看ACL
cloud-storage-tool s3 acl data-bucket get --format json
```

##### 6. 桶清单配置 (Bucket Inventory)

**命令格式**
```bash
cloud-storage-tool <provider> inventory <bucket-name> <action> [options]
```

**子命令**
| 子命令 | 说明 |
|--------|------|
| `create` | 创建清单配置 |
| `list` | 列出所有清单配置 |
| `get` | 获取特定清单配置 |
| `update` | 更新清单配置 |
| `delete` | 删除清单配置 |
| `enable` | 启用清单配置 |
| `disable` | 禁用清单配置 |

**create子命令参数**
| 参数 | 缩写 | 说明 | 默认值 |
|------|------|------|--------|
| `--name` | `-n` | 清单配置名称 | 必需 |
| `--destination` | `-d` | 目标桶（存储清单报告的桶） | 必需 |
| `--prefix` | `-p` | 清单报告存储前缀 | inventory/ |
| `--format` | `-f` | 报告格式（CSV, Parquet, ORC） | CSV |
| `--schedule` | `-s` | 生成频率（Daily, Weekly） | Daily |
| `--included-objects` | `-i` | 包含的对象前缀（多个用逗号分隔） | 所有对象 |
| `--excluded-objects` | `-e` | 排除的对象前缀（多个用逗号分隔） | 无 |
| `--optional-fields` | `-o` | 可选字段（Size, LastModifiedDate等） | 基础字段 |

**其他通用参数**
| 参数 | 缩写 | 说明 | 默认值 |
|------|------|------|--------|
| `--region` | `-r` | 桶所在的区域 | 自动检测 |
| `--config-id` | `-c` | 清单配置ID（用于get/update/delete） | 必需 |

**使用示例**
```bash
# 创建每日清单配置
cloud-storage-tool cos inventory data-bucket create \
  --name daily-inventory \
  --destination report-bucket \
  --schedule Daily \
  --format CSV

# 列出所有清单配置
cloud-storage-tool s3 inventory logs-bucket list

# 获取特定清单配置详情
cloud-storage-tool cos inventory backup-bucket get \
  --config-id config-123

# 更新清单配置，增加可选字段
cloud-storage-tool s3 inventory data-bucket update \
  --config-id config-456 \
  --optional-fields Size,LastModifiedDate,StorageClass

# 删除清单配置
cloud-storage-tool cos inventory temp-bucket delete \
  --config-id config-789

# 创建包含特定对象的周报
cloud-storage-tool s3 inventory app-bucket create \
  --name weekly-app-inventory \
  --destination analytics-bucket \
  --schedule Weekly \
  --included-objects "app/", "logs/" \
  --format Parquet
```

##### 7. 生命周期规则管理 (Lifecycle Rules)

**命令格式**
```bash
cloud-storage-tool <provider> lifecycle <bucket-name> <action> [options]
```

**子命令**
| 子命令 | 说明 |
|--------|------|
| `create` | 创建生命周期规则 |
| `list` | 列出所有生命周期规则 |
| `get` | 获取特定规则详情 |
| `update` | 更新生命周期规则 |
| `delete` | 删除生命周期规则 |

**create子命令参数**
| 参数 | 缩写 | 说明 | 默认值 |
|------|------|------|--------|
| `--id` | `-i` | 规则ID | 必需 |
| `--prefix` | `-p` | 规则适用的对象前缀 | 所有对象 |
| `--status` | `-s` | 规则状态（Enabled, Disabled） | Enabled |
| `--transition-days` | `-t` | 转换到低频存储的天数 | 无 |
| `--transition-storage-class` | `-c` | 转换后的存储类型 | STANDARD_IA |
| `--expiration-days` | `-e` | 过期删除的天数 | 无 |
| `--abort-incomplete-multipart-days` | `-a` | 清理未完成分块上传的天数 | 无 |
| `--noncurrent-version-transition-days` | `-n` | 非当前版本转换天数 | 无 |
| `--noncurrent-version-expiration-days` | `-x` | 非当前版本过期天数 | 无 |
| `--time-type` | `-T` | 时间类型（mtime:修改时间, atime:访问时间） | mtime |

**其他通用参数**
| 参数 | 缩写 | 说明 | 默认值 |
|------|------|------|--------|
| `--region` | `-r` | 桶所在的区域 | 自动检测 |
| `--rule-id` | `-r` | 规则ID（用于get/update/delete） | 必需 |

**生命周期规则时间类型说明**

##### 1. 基于修改时间 (mtime) 的生命周期规则
- **定义**：基于对象的最后修改时间计算生命周期
- **支持范围**：所有云服务提供商都支持
- **适用场景**：常规对象管理，如日志文件、备份文件等
- **支持所有生命周期参数**

##### 2. 基于访问时间 (atime) 的生命周期规则
- **定义**：基于对象的最后访问时间计算生命周期
- **支持范围**：有限支持，需要特定功能开启
- **适用场景**：智能分层存储，根据访问频率优化存储成本

##### 3. 云服务商支持情况

**AWS S3 支持情况：**
- ✅ **mtime规则**：完全支持所有参数
- ✅ **atime规则**：通过S3 Intelligent-Tiering支持
  - 需要启用S3 Intelligent-Tiering存储类型
  - 支持自动分层优化，基于访问模式
  - **atime规则限制**：
    - 不支持`--noncurrent-version-transition-days`
    - 不支持`--noncurrent-version-expiration-days`
    - 不支持`--abort-incomplete-multipart-days`
    - 转换天数通常固定（如30、90、180、365天）

**腾讯云COS 支持情况：**
- ✅ **mtime规则**：完全支持所有参数
- ⚠️ **atime规则**：有限支持或需要特殊配置
  - 可能需要联系技术支持开启
  - 支持程度可能因区域而异
  - **建议优先使用mtime规则**

##### 4. atime生命周期规则参数限制
以下参数在atime生命周期规则中**可能不支持或有限制**：
- `--noncurrent-version-transition-days`：非当前版本转换
- `--noncurrent-version-expiration-days`：非当前版本过期
- `--abort-incomplete-multipart-days`：清理未完成分块上传
- `--transition-storage-class`：转换目标存储类型可能受限

**使用建议：**
1. 默认使用`--time-type mtime`（修改时间规则）
2. 只有在需要智能分层时才使用`--time-type atime`
3. 使用atime规则前，确认云服务商支持情况
4. 对于重要数据，建议先进行小规模测试

**使用示例**
```bash
# 创建基于修改时间（mtime）的规则：30天后转为低频存储
cloud-storage-tool cos lifecycle data-bucket create \
  --id transition-to-ia \
  --prefix "logs/" \
  --time-type mtime \
  --transition-days 30 \
  --transition-storage-class STANDARD_IA

# 创建基于访问时间（atime）的智能分层规则（仅S3支持）
cloud-storage-tool s3 lifecycle archive-bucket create \
  --id smart-tiering \
  --prefix "archive/" \
  --time-type atime \
  --transition-days 90

# 创建90天后删除的规则（基于修改时间）
cloud-storage-tool s3 lifecycle temp-bucket create \
  --id delete-after-90d \
  --prefix "temp/" \
  --time-type mtime \
  --expiration-days 90

# 列出所有生命周期规则
cloud-storage-tool cos lifecycle archive-bucket list

# 获取规则详情
cloud-storage-tool s3 lifecycle data-bucket get \
  --rule-id rule-123

# 更新规则
cloud-storage-tool cos lifecycle logs-bucket update \
  --rule-id old-rule \
  --expiration-days 180 \
  --status Enabled

# 删除规则
cloud-storage-tool s3 lifecycle test-bucket delete \
  --rule-id unused-rule

# 创建复杂规则：7天转低频，365天删除（基于修改时间）
cloud-storage-tool cos lifecycle backup-bucket create \
  --id comprehensive-rule \
  --status Enabled \
  --time-type mtime \
  --transition-days 7 \
  --transition-storage-class STANDARD_IA \
  --expiration-days 365
```

##### 2.2.4 对象操作详细参考

对象操作支持统一的语法，使用URI格式指定对象位置：
- **本地文件**: `local/path/to/file.txt`
- **COS对象**: `cos://bucket-name/path/to/object`
- **S3对象**: `s3://bucket-name/path/to/object`

###### 2.2.4.1 对象上传 (Upload)

**命令格式**
```bash
cloud-storage-tool cp <source> <destination> [options]
# 或使用明确的上传命令
cloud-storage-tool upload <local-file> <cloud-object> [options]
```

**参数说明**
| 参数 | 缩写 | 说明 | 默认值 |
|------|------|------|--------|
| `--recursive` | `-r` | 递归上传目录 | false |
| `--exclude` | `-e` | 排除模式（支持通配符） | 无 |
| `--include` | `-i` | 包含模式（支持通配符） | 所有文件 |
| `--storage-class` | `-s` | 存储类型（STANDARD/STANDARD_IA/ARCHIVE等） | STANDARD |
| `--metadata` | `-m` | 自定义元数据（key=value格式） | 无 |
| `--acl` | `-a` | 对象访问权限 | bucket默认 |
| `--part-size` | `-p` | 分块上传大小（字节） | 10MB |
| `--threads` | `-t` | 并发上传线程数 | 10 |
| `--dry-run` | `-d` | 模拟运行，不上传实际文件 | false |
| `--checksum` | `-c` | 启用完整性校验（MD5/SHA256） | true |
| `--resume` | `-R` | 支持断点续传 | true |

**使用示例**
```bash
# 上传单个文件到COS
cloud-storage-tool cp local/file.txt cos://my-bucket/path/file.txt

# 上传单个文件到S3
cloud-storage-tool cp local/file.txt s3://my-bucket/path/file.txt

# 递归上传目录到COS
cloud-storage-tool cp local/directory/ cos://my-bucket/backup/ --recursive

# 上传并设置存储类型为低频访问
cloud-storage-tool cp data.log cos://logs-bucket/app/data.log \
  --storage-class STANDARD_IA

# 上传并设置自定义元数据
cloud-storage-tool cp image.jpg s3://media-bucket/images/image.jpg \
  --metadata "author=john,project=website"

# 排除特定文件上传
cloud-storage-tool cp source/ cos://backup-bucket/source/ \
  --recursive \
  --exclude "*.tmp" \
  --exclude ".git/*"

# 大文件分块上传（100MB分块）
cloud-storage-tool cp large-file.iso cos://backup-bucket/large-file.iso \
  --part-size 104857600 \
  --threads 20

# 模拟上传（不实际执行）
cloud-storage-tool cp local/dir/ s3://bucket/dir/ --recursive --dry-run
```

###### 2.2.4.2 对象下载 (Download)

**命令格式**
```bash
cloud-storage-tool cp <cloud-object> <local-destination> [options]
# 或使用明确的下载命令
cloud-storage-tool download <cloud-object> <local-file> [options]
```

**参数说明**
| 参数 | 缩写 | 说明 | 默认值 |
|------|------|------|--------|
| `--recursive` | `-r` | 递归下载目录 | false |
| `--exclude` | `-e` | 排除模式（支持通配符） | 无 |
| `--include` | `-i` | 包含模式（支持通配符） | 所有对象 |
| `--part-size` | `-p` | 分块下载大小（字节） | 10MB |
| `--threads` | `-t` | 并发下载线程数 | 10 |
| `--dry-run` | `-d` | 模拟运行，不下载实际文件 | false |
| `--checksum` | `-c` | 启用完整性校验 | true |
| `--resume` | `-R` | 支持断点续传 | true |
| `--force` | `-f` | 强制覆盖已存在的本地文件 | false |
| `--latest` | `-l` | 只下载最新版本（版本控制桶） | false |

**使用示例**
```bash
# 下载单个文件从COS
cloud-storage-tool cp cos://my-bucket/path/file.txt local/file.txt

# 下载单个文件从S3
cloud-storage-tool cp s3://my-bucket/path/file.txt local/file.txt

# 递归下载目录从COS
cloud-storage-tool cp cos://my-bucket/backup/ local/backup/ --recursive

# 下载特定前缀的对象
cloud-storage-tool cp s3://logs-bucket/app/ local/logs/ \
  --recursive \
  --include "*.log" \
  --exclude "*.tmp"

# 大文件分块下载
cloud-storage-tool cp cos://backup-bucket/large-file.iso local/large-file.iso \
  --part-size 104857600 \
  --threads 20

# 强制覆盖本地文件
cloud-storage-tool cp s3://bucket/file.txt local/file.txt --force

# 只下载最新版本
cloud-storage-tool cp cos://versioned-bucket/doc.pdf local/doc.pdf --latest

# 模拟下载（不实际执行）
cloud-storage-tool cp cos://bucket/dir/ local/dir/ --recursive --dry-run
```

###### 2.2.4.3 对象删除 (Delete)

**命令格式**
```bash
cloud-storage-tool rm <cloud-object> [options]
```

**参数说明**
| 参数 | 缩写 | 说明 | 默认值 |
|------|------|------|--------|
| `--recursive` | `-r` | 递归删除目录 | false |
| `--exclude` | `-e` | 排除模式（支持通配符） | 无 |
| `--include` | `-i` | 包含模式（支持通配符） | 所有对象 |
| `--dry-run` | `-d` | 模拟运行，不实际删除 | false |
| `--force` | `-f` | 跳过确认提示 | false |
| `--versions` | `-v` | 删除所有版本（版本控制桶） | false |
| `--older-than` | `-o` | 只删除早于指定天数的对象 | 无 |
| `--prefix` | `-p` | 删除指定前缀的所有对象 | 无 |

**使用示例**
```bash
# 删除单个对象
cloud-storage-tool rm cos://my-bucket/path/file.txt

# 删除S3对象
cloud-storage-tool rm s3://my-bucket/path/file.txt

# 递归删除目录（需要确认）
cloud-storage-tool rm cos://my-bucket/old-data/ --recursive

# 强制删除，跳过确认
cloud-storage-tool rm s3://logs-bucket/temp/ --recursive --force

# 删除匹配特定模式的对象
cloud-storage-tool rm cos://bucket/data/ \
  --recursive \
  --include "*.tmp" \
  --exclude "important.tmp"

# 删除所有版本（版本控制桶）
cloud-storage-tool rm s3://versioned-bucket/document.txt --versions

# 删除30天前的旧文件
cloud-storage-tool rm cos://backup-bucket/ \
  --recursive \
  --older-than 30 \
  --include "*.log"

# 删除指定前缀的所有对象
cloud-storage-tool rm s3://bucket/ --prefix "temp/"

# 模拟删除（不实际执行）
cloud-storage-tool rm cos://bucket/to-delete/ --recursive --dry-run
```

###### 2.2.4.4 对象列表 (List)

**命令格式**
```bash
cloud-storage-tool ls <cloud-path> [options]
```

**参数说明**
| 参数 | 缩写 | 说明 | 默认值 |
|------|------|------|--------|
| `--recursive` | `-r` | 递归列出所有对象 | false |
| `--human-readable` | `-h` | 人类可读的文件大小 | false |
| `--long` | `-l` | 长格式显示（详细信息） | false |
| `--all` | `-a` | 显示所有对象（包括删除标记） | false |
| `--versions` | `-v` | 显示所有版本（版本控制桶） | false |
| `--limit` | `-n` | 最多显示的对象数量 | 无限制 |
| `--marker` | `-m` | 起始标记（分页） | 无 |
| `--prefix` | `-p` | 只显示指定前缀的对象 | 无 |
| `--delimiter` | `-d` | 目录分隔符 | / |
| `--format` | `-f` | 输出格式（json, table, csv） | table |

**使用示例**
```bash
# 列出桶根目录
cloud-storage-tool ls cos://my-bucket/

# 列出S3桶目录
cloud-storage-tool ls s3://my-bucket/path/

# 递归列出所有对象
cloud-storage-tool ls cos://backup-bucket/ --recursive

# 长格式显示详细信息
cloud-storage-tool ls s3://logs-bucket/ --long --human-readable

# 显示所有版本（版本控制桶）
cloud-storage-tool ls cos://versioned-bucket/ --versions

# 只显示特定前缀的对象
cloud-storage-tool ls s3://bucket/ --prefix "images/"

# 分页列出对象
cloud-storage-tool ls cos://large-bucket/ --limit 100
cloud-storage-tool ls cos://large-bucket/ --limit 100 --marker "last-object-key"

# JSON格式输出
cloud-storage-tool ls s3://bucket/ --format json

# 显示目录结构（使用分隔符）
cloud-storage-tool ls cos://bucket/ --delimiter "/"
```

###### 2.2.4.5 对象信息查看 (Stat)

**命令格式**
```bash
cloud-storage-tool stat <cloud-object> [options]
```

**参数说明**
| 参数 | 缩写 | 说明 | 默认值 |
|------|------|------|--------|
| `--version-id` | `-v` | 指定版本ID（版本控制桶） | 最新版本 |
| `--format` | `-f` | 输出格式（json, yaml, table） | table |
| `--metadata-only` | `-m` | 只显示元数据 | false |
| `--checksum` | `-c` | 包含校验和信息 | false |

**显示信息包括**
- 对象键、大小、最后修改时间
- 存储类型、ETag、版本ID
- 自定义元数据
- 访问权限
- 加密信息（如果已加密）

**使用示例**
```bash
# 查看对象基本信息
cloud-storage-tool stat cos://my-bucket/path/file.txt

# 查看S3对象信息
cloud-storage-tool stat s3://my-bucket/path/file.txt

# 查看特定版本的对象
cloud-storage-tool stat cos://versioned-bucket/doc.pdf --version-id "abc123"

# JSON格式输出
cloud-storage-tool stat s3://bucket/object --format json

# 只显示元数据
cloud-storage-tool stat cos://bucket/object --metadata-only

# 包含校验和信息
cloud-storage-tool stat s3://bucket/object --checksum
```

###### 2.2.4.6 生成预签名URL (Presign)

**命令格式**
```bash
cloud-storage-tool presign <cloud-object> [options]
```

**参数说明**
| 参数 | 缩写 | 说明 | 默认值 |
|------|------|------|--------|
| `--expires` | `-e` | URL有效期（秒） | 3600 |
| `--method` | `-m` | HTTP方法（GET, PUT, DELETE） | GET |
| `--response-headers` | `-r` | 响应头设置 | 无 |
| `--version-id` | `-v` | 指定版本ID（版本控制桶） | 最新版本 |
| `--content-type` | `-c` | 指定Content-Type（PUT方法） | 自动检测 |

**使用示例**
```bash
# 生成1小时有效的下载URL
cloud-storage-tool presign cos://my-bucket/path/file.txt

# 生成24小时有效的S3下载URL
cloud-storage-tool presign s3://my-bucket/path/file.txt --expires 86400

# 生成上传URL（PUT方法）
cloud-storage-tool presign s3://bucket/upload-target \
  --method PUT \
  --expires 1800

# 生成带响应头控制的URL
cloud-storage-tool presign cos://bucket/image.jpg \
  --response-headers "response-content-type=image/jpeg" \
  --response-headers "response-content-disposition=attachment"

# 生成特定版本对象的URL
cloud-storage-tool presign s3://versioned-bucket/doc.pdf \
  --version-id "xyz789" \
  --expires 7200

# 生成带Content-Type的上传URL
cloud-storage-tool presign cos://bucket/upload.json \
  --method PUT \
  --content-type "application/json" \
  --expires 3600
```

###### 2.2.4.7 复制对象 (Copy)

**命令格式**
```bash
cloud-storage-tool cp <source> <destination> [options]
```

**参数说明**
| 参数 | 缩写 | 说明 | 默认值 |
|------|------|------|--------|
| `--recursive` | `-r` | 递归复制目录 | false |
| `--metadata` | `-m` | 复制或替换元数据 | 继承源对象 |
| `--storage-class` | `-s` | 目标存储类型 | 继承源对象 |
| `--acl` | `-a` | 目标对象访问权限 | 继承源对象 |
| `--dry-run` | `-d` | 模拟运行，不实际复制 | false |
| `--cross-provider` | `-x` | 跨提供商复制（COS ↔ S3） | 自动检测 |

**使用示例**
```bash
# 桶内复制对象
cloud-storage-tool cp cos://bucket/src/file.txt cos://bucket/dst/file.txt

# 跨桶复制
cloud-storage-tool cp s3://source-bucket/data.log s3://dest-bucket/backup/data.log

# 递归复制目录
cloud-storage-tool cp cos://bucket/source/ cos://bucket/destination/ --recursive

# 跨提供商复制（COS到S3）
cloud-storage-tool cp cos://cos-bucket/data s3://s3-bucket/data --cross-provider

# 复制并更改存储类型
cloud-storage-tool cp s3://bucket/hot-data s3://bucket/cold-data \
  --storage-class GLACIER

# 复制并设置新元数据
cloud-storage-tool cp cos://bucket/original cos://bucket/copy \
  --metadata "copied=true,timestamp=$(date +%s)"

# 模拟复制
cloud-storage-tool cp cos://src/dir/ cos://dst/dir/ --recursive --dry-run
```

###### 2.2.4.8 移动对象 (Move)

**命令格式**
```bash
cloud-storage-tool mv <source> <destination> [options]
```

**参数说明**
| 参数 | 缩写 | 说明 | 默认值 |
|------|------|------|--------|
| `--recursive` | `-r` | 递归移动目录 | false |
| `--dry-run` | `-d` | 模拟运行，不实际移动 | false |
| `--force` | `-f` | 强制覆盖目标对象 | false |
| `--cross-provider` | `-x` | 跨提供商移动（COS ↔ S3） | 自动检测 |

**注意**：移动操作实际是复制+删除，原子性不保证。对于重要数据，建议先复制再删除。

**使用示例**
```bash
# 移动单个对象
cloud-storage-tool mv cos://bucket/old/path.txt cos://bucket/new/path.txt

# 移动S3对象
cloud-storage-tool mv s3://bucket/source.log s3://bucket/archived/source.log

# 递归移动目录
cloud-storage-tool mv cos://bucket/temp/ cos://bucket/archive/ --recursive

# 跨提供商移动
cloud-storage-tool mv cos://cos-bucket/data s3://s3-bucket/data --cross-provider

# 强制覆盖目标
cloud-storage-tool mv s3://bucket/newer s3://bucket/older --force

# 模拟移动
cloud-storage-tool mv cos://src/dir/ cos://dst/dir/ --recursive --dry-run
```



##### 2.2.5 全局选项

**通用选项（适用于所有命令）**
| 选项 | 缩写 | 说明 | 默认值 |
|------|------|------|--------|
| `--config` | `-c` | 指定配置文件路径 | ~/.cloud-storage/config.yaml |
| `--profile` | `-p` | 使用指定的配置profile | default |
| `--debug` | `-d` | 启用调试模式 | false |
| `--quiet` | `-q` | 安静模式，只输出必要信息 | false |
| `--help` | `-h` | 显示命令帮助 | false |
| `--version` | `-v` | 显示版本信息 | false |

##### 2.2.6 环境变量配置

**认证信息（必须通过环境变量设置）**
```bash
# 腾讯云COS
export TENCENT_COS_SECRET_ID="your-secret-id"
export TENCENT_COS_SECRET_KEY="your-secret-key"

# AWS S3
export AWS_ACCESS_KEY_ID="your-access-key-id"
export AWS_SECRET_ACCESS_KEY="your-secret-access-key"
export AWS_REGION="ap-singapore"  # 可选，可覆盖配置文件
```



## 3. 非功能需求

### 3.1 性能要求
- **上传速度**：支持多线程并发上传
- **大文件支持**：支持GB级别大文件
- **内存占用**：流式处理，避免大内存占用
- **网络优化**：支持断点续传、重试机制

### 3.2 可靠性要求
- **错误处理**：完善的错误处理和恢复机制
- **日志记录**：详细的操作日志和错误日志
- **数据完整性**：上传下载校验（MD5/SHA256）
- **事务性**：重要操作支持回滚

### 3.3 安全性要求
- **密钥管理**：支持环境变量、配置文件、密钥管理服务
- **最小权限**：遵循最小权限原则
- **审计日志**：所有操作记录审计日志
- **数据加密**：支持客户端和服务端加密

### 3.4 易用性要求
- **直观的命令**：类似aws s3、coscmd的熟悉语法
- **详细的帮助**：完整的帮助文档和错误提示
- **进度显示**：上传下载进度条显示
- **配置简化**：一键配置和测试

## 4. 技术架构

### 4.1 技术栈
- **编程语言**：Go 1.21+
- **云服务SDK**：
  - 腾讯云COS：`github.com/tencentyun/cos-go-sdk-v5`
  - AWS S3：`github.com/aws/aws-sdk-go-v2/service/s3`
- **命令行框架**：Cobra + Viper
- **配置文件**：YAML格式
- **测试框架**：Go test + 模拟测试

### 4.2 架构设计
```
cloud-storage-tool/
├── cmd/                    # 命令行入口
├── internal/
│   ├── provider/          # 云服务提供商接口
│   │   ├── cos.go         # 腾讯云COS实现
│   │   ├── s3.go          # AWS S3实现
│   │   └── interface.go   # 统一接口定义
│   ├── commands/          # 具体命令实现
│   │   ├── bucket/        # 桶操作命令
│   │   ├── object/        # 对象操作命令
│   │   └── sync/          # 同步命令
│   ├── config/            # 配置管理
│   └── utils/             # 工具函数
├── pkg/                   # 公共包
├── config/                # 配置文件示例
├── scripts/               # 构建和部署脚本
└── docs/                  # 文档
```

### 4.3 统一接口设计
```go
// 存储提供商统一接口
type StorageProvider interface {
    // 桶操作
    ListBuckets(ctx context.Context) ([]Bucket, error)
    CreateBucket(ctx context.Context, name, region string) error
    DeleteBucket(ctx context.Context, name string, force bool) error
    
    // 对象操作
    ListObjects(ctx context.Context, bucket, prefix string) ([]Object, error)
    UploadFile(ctx context.Context, bucket, key, filepath string) error
    DownloadFile(ctx context.Context, bucket, key, filepath string) error
    DeleteObject(ctx context.Context, bucket, key string) error
    
    // 高级功能
    CopyObject(ctx context.Context, src, dst ObjectInfo) error
    GetPresignedURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error)
}
```

## 5. 配置管理

### 5.1 配置文件格式
```yaml
# ~/.cloud-storage/config.yaml
# 安全提示：敏感信息（AK/SK）不应明文存储在配置文件中
# 请使用环境变量进行配置

providers:
  cos:
    enabled: true
    # 必须通过环境变量配置，不在配置文件中明文存储
    secret_id: ${TENCENT_COS_SECRET_ID}
    secret_key: ${TENCENT_COS_SECRET_KEY}
    default_region: ap-singapore
    default_bucket: openclaw-bakup-1251036673
    endpoint: ""  # 可选，自定义端点
    
  s3:
    enabled: true
    # 必须通过环境变量配置，不在配置文件中明文存储
    access_key: ${AWS_ACCESS_KEY_ID}
    secret_key: ${AWS_SECRET_ACCESS_KEY}
    region: ${AWS_REGION}
    endpoint: ""  # 可选，用于兼容S3协议的其他服务
    force_path_style: false

defaults:
  provider: cos    # 默认提供商
  threads: 10      # 默认并发数
  retry_count: 3   # 默认重试次数
  timeout: 300     # 默认超时时间（秒）
  part_size: 10485760    # 分块上传大小（10MB）
  
# 同步设置
sync:
  exclude_patterns:      # 排除模式（类似.gitignore）
    - "*.tmp"
    - "*.log"
    - ".git/*"
    - ".DS_Store"
  compare_method: mtime  # 同步比较方法：mtime|size|checksum
  delete: false          # 是否删除目标端多余文件
  
# 日志设置
logging:
  level: info           # 日志级别：debug|info|warn|error
  format: text          # 日志格式：text|json
  file: ""              # 日志文件路径（空则输出到控制台）
```

### 5.2 配置来源优先级
1. 命令行参数（最高优先级）
2. 环境变量
3. 配置文件
4. 默认值（最低优先级）

## 6. 项目计划

### 6.1 开发阶段

#### 第一阶段：基础功能（2周）
- 项目初始化、架构搭建
- 配置管理系统
- 基础命令框架
- COS提供商基本实现

#### 第二阶段：核心功能（3周）
- S3提供商基本实现
- 桶管理命令实现
- 对象管理命令实现
- 错误处理和日志系统

#### 第三阶段：高级功能（2周）
- 同步功能实现
- 批量操作优化
- 性能优化和测试

#### 第四阶段：完善和文档（1周）
- 完整测试覆盖
- 文档编写
- 性能优化
- 发布准备

### 6.2 里程碑
- **M1**：基础架构完成，支持COS基本操作
- **M2**：支持S3基本操作，桶和对象管理
- **M3**：同步功能完成，性能优化
- **M4**：完整测试，文档完成，正式发布

## 7. 成功标准

### 7.1 功能标准
- 支持腾讯云COS所有核心操作
- 支持AWS S3所有核心操作
- 提供统一的命令行接口
- 支持配置管理和环境变量

### 7.2 性能标准
- 上传下载速度不低于原生SDK的90%
- 内存占用控制在合理范围
- 支持大文件断点续传

### 7.3 质量标准
- 测试覆盖率 > 80%
- 完整的文档和示例
- 易于安装和配置

## 8. 风险与缓解

### 8.1 技术风险
- **SDK兼容性**：不同版本SDK API变化
  - 缓解：使用稳定的SDK版本，定期更新测试
- **跨云差异**：COS和S3功能差异
  - 缓解：统一接口设计，差异功能单独处理

### 8.2 安全风险
- **密钥泄露**：配置文件中的敏感信息
  - 缓解：支持密钥管理服务，环境变量优先

### 8.3 运营风险
- **维护成本**：需要持续跟进云服务商API变化
  - 缓解：模块化设计，易于扩展新提供商

## 9. 扩展计划

- 支持阿里云OSS
- 支持华为云OBS
- 支持Google Cloud Storage

---

## 下一步行动建议

1. **确认需求优先级**：哪些功能是最急需的？
2. **设计评审**：技术架构是否合理？
3. **开始实施**：从哪个模块开始开发？
4. **测试计划**：如何验证各功能模块？

---

**文档版本**：1.6  
**创建日期**：2026-03-05  
**最后更新**：2026-03-05  
**状态**：草案