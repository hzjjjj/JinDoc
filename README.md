类似百度网盘的家庭文件存储交换网站，可以部署在家庭电脑，机顶盒或NAS和服务器上。后端Go语言，前端纯javascript。数据库请用.env定义redis或upstash无服务redis数据库

#Jin多媒体分享平台创作过程

#### ===================【安装】========================
上传到linux
nohup ./jindoc_0.38 > nohup_jindoc38.log 2>&1 &

#### ===================【.env文件】========================

```
# Jin 家庭_办公室_团队 多媒体分享平台配置中心
# 以下设置不支持中文

# 网站标题因UTF-8编码乱码问题不支持中文
APP_TITLE = "NetFlix2025"

# 端口
PORT = 8800

# 生产部署和开发模式"production"和"dev"
ENV = "dev"

# 上传文章和文档过期时间（小时）
EXPIRATION_TIME = 480

# Redis数据库地址，可含密码和IP地址
# 旧参数UPSTASH_REDIS_REST_URL = "rediss://你的数据库网址和端口:6379"
UPSTASH_REDIS_REST_URL = "rediss://default:你的API_token@数据库网址:6379"

# 测试杂项
UPSTASH_REDIS_REST_TOKEN = "你的API_token"
UPSTASH_REDIS_URL = "https://你的数据库网址"
```

#### ===================【设计原则】========================

1、解决家庭，工作单位，团队，亲友之间分享文章，图片和文件。
2、支持各类手机和电脑PC端之间传输，轻量高效。
3、弥补微信钉钉，各类网盘传输分享上的不足：过程复杂，要下载安装APP，加好友和充会员，网络审查不够隐私。

#### ===================【待解决功能】========================

文章和文档显示到期时间和倒计时
可以删除文章，删除文档和照片时可以同时删除网站对应文件。
过期文件和文章自动转到回收站，30天删除
多媒体上传页面布局和美化
阅后即焚：1小时删除或浏览后自动删除
全站用svelet框架美化并提供交互功能
待解决go语言代码中的收藏图标语句可以删除
所有图片放到img文件夹中，不用放到uploads中

#### ===================【已解决问题】========================

统一env环境变量设置和读取，包括数据库地址，网站标题，上传文章和文档的7天删除
上传文章7天删除★
照片列表不显示照片预览
文章列表显示文章内容预览
列表显示最新上传的内容
优化redis数据储存逻辑，便于提取最新文章列表。
文章全文阅读按钮失效和模态框或展开阅读方式尝试。
解决数据库读取过多问题，比如打开首页要读取10次数据库。rd2-read-get  
//收藏夹图标 ico 待解决go语言代码中的收藏图标语句可以删除

### ===================【设置环境变量】========================

怎么设置环境变量：UPSTASH_REDIS_URL和UPSTASH_REDIS_REST_TOKEN

设置环境变量的方法取决于你使用的操作系统和开发环境。以下是针对不同场景的设置方法：

1. 临时设置（当前终端会话有效）
Linux/macOS
在终端中执行以下命令：

export UPSTASH_REDIS_URL="your-redis-url-here"
export UPSTASH_REDIS_REST_TOKEN="your-token-here"
Windows（PowerShell）
$env:UPSTASH_REDIS_URL="your-redis-url-here"
$env:UPSTASH_REDIS_REST_TOKEN="your-token-here"
Windows（命令提示符）
set UPSTASH_REDIS_URL=your-redis-url-here
set UPSTASH_REDIS_REST_TOKEN=your-token-here
2. 永久设置（所有终端会话有效）
Linux/macOS
编辑 ~/.bashrc 或 ~/.zshrc 文件（根据你使用的 shell）：

echo 'export UPSTASH_REDIS_URL="your-redis-url-here"' >> ~/.bashrc
echo 'export UPSTASH_REDIS_REST_TOKEN="your-token-here"' >> ~/.bashrc
source ~/.bashrc  # 立即生效
Windows（系统环境变量）
搜索并打开 "系统属性" → "高级系统设置" → "环境变量"。
在 "系统变量" 区域点击 "新建"，分别添加 UPSTASH_REDIS_URL 和 UPSTASH_REDIS_REST_TOKEN。
重启终端或应用程序使设置生效。
3. 在IDE或编辑器中设置（以VS Code为例）
如果你使用 VS Code 开发，可以在项目根目录创建 .env 文件（需安装 dotenv 包）：

UPSTASH_REDIS_URL=your-redis-url-here
UPSTASH_REDIS_REST_TOKEN=your-token-here
