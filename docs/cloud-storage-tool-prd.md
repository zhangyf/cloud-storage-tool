# 云存储统一管理工具 PRD (Product Requirements Document)

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

#### 2.1.4 高级功能
- **数据迁移**：COS ↔ S3 数据迁移
- **存储分析**：分析存储使用情况
- **生命周期管理**：管理对象过期策略
- **版本控制**：支持版本化桶的操作

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
providers:
  cos:
    enabled: true
    secret_id: ${TENCENT_COS_SECRET_ID}
    secret_key: ${TENCENT_COS_SECRET_KEY}
    default_region: ap-singapore
    default_bucket: openclaw-bakup-1251036673
    
  s3:
    enabled: true
    access_key: ${AWS_ACCESS_KEY_ID}
    secret_key: ${AWS_SECRET_ACCESS_KEY}
    region: ${AWS_REGION}
    endpoint: ""  # 可选，用于兼容S3协议的其他服务

defaults:
  provider: cos    # 默认提供商
  threads: 10      # 默认并发数
  retry_count: 3   # 默认重试次数
  timeout: 300     # 默认超时时间（秒）

logging:
  level: info
  format: json
  file: /var/log/cloud-storage.log
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

## 9. 后续扩展

### 9.1 短期扩展
- 支持阿里云OSS
- 支持华为云OBS
- 支持Google Cloud Storage

### 9.2 长期愿景
- Web管理界面
- 图形化客户端
- RESTful API服务
- 存储成本分析和优化建议

---

## 下一步行动建议

1. **确认需求优先级**：哪些功能是最急需的？
2. **设计评审**：技术架构是否合理？
3. **开始实施**：从哪个模块开始开发？
4. **测试计划**：如何验证各功能模块？

---

**文档版本**：1.0  
**创建日期**：2026-03-05  
**最后更新**：2026-03-05  
**状态**：草案