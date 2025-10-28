# 🎨 Hướng Dẫn Tùy Chỉnh Giao Diện Blockscout

## 📋 Tổng Quan

Hướng dẫn này sẽ giúp bạn tùy chỉnh giao diện Blockscout để:
- ✅ Thay logo và branding
- ✅ Loại bỏ quảng cáo
- ✅ Tùy chỉnh màu sắc và theme
- ✅ Thay đổi layout và styling

## 🎯 Các File Cần Chỉnh Sửa

### 1. **Logo và Branding**

#### 📁 Thay thế logo chính:
```
📂 apps/block_scout_web/assets/static/images/
├── 🖼️ logo.svg (logo chính)
├── 🖼️ logo-light.svg (logo cho dark mode)
├── 🖼️ favicon.ico (icon trình duyệt)
└── 🖼️ apple-touch-icon.png (icon iOS)
```

#### 📁 Tạo logo mới:
1. Tạo file `logo.svg` với kích thước tối ưu (200x50px)
2. Tạo file `logo-light.svg` cho dark mode
3. Tạo file `favicon.ico` (32x32px hoặc 16x16px)

### 2. **Theme và Màu Sắc**

#### 📁 File theme chính:
```
📂 apps/block_scout_web/assets/css/theme/
├── 🎨 poatc-theme.scss (theme POATC hiện tại)
├── 🎨 _variables.scss (biến màu sắc)
└── 🎨 _colors.scss (định nghĩa màu)
```

### 3. **Loại Bỏ Quảng Cáo**

#### 📁 Files quảng cáo cần disable:
```
📂 apps/block_scout_web/assets/js/lib/
├── 🚫 ad.js (quảng cáo chính)
├── 🚫 banner.js (banner ads)
├── 🚫 text_ad.js (text ads)
└── 🚫 custom_ad.json (custom ads config)
```

## 🛠️ Các Bước Thực Hiện

### Bước 1: Thay Logo

#### 1.1. Tạo logo mới:
```bash
# Di chuyển đến thư mục images
cd apps/block_scout_web/assets/static/images/

# Backup logo cũ
cp logo.svg logo.svg.backup
cp logo-light.svg logo-light.svg.backup
cp favicon.ico favicon.ico.backup
```

#### 1.2. Thay thế logo:
- Đặt logo mới vào thư mục `images/`
- Đảm bảo tên file giống với file cũ
- Kiểm tra kích thước phù hợp

### Bước 2: Tùy Chỉnh Theme

#### 2.1. Chỉnh sửa màu sắc trong `poatc-theme.scss`:
```scss
// apps/block_scout_web/assets/css/theme/poatc-theme.scss

// Màu chính
$primary-color: #your-color;
$secondary-color: #your-color;

// Màu nền
$background-color: #your-color;
$card-background: #your-color;

// Màu text
$text-color: #your-color;
$text-muted: #your-color;

// Màu border
$border-color: #your-color;

// Màu button
$button-primary: #your-color;
$button-secondary: #your-color;
```

#### 2.2. Tùy chỉnh layout:
```scss
// Thay đổi padding/margin
$container-padding: 20px;
$card-padding: 15px;

// Thay đổi border radius
$border-radius: 8px;
$button-border-radius: 6px;

// Thay đổi font
$font-family: 'Your-Font', sans-serif;
```

### Bước 3: Loại Bỏ Quảng Cáo

#### 3.1. Disable ads trong `ad.js`:
```javascript
// apps/block_scout_web/assets/js/lib/ad.js

function showAd () {
  // Luôn trả về false để disable ads
  return false;
}
```

#### 3.2. Disable banner ads trong `banner.js`:
```javascript
// apps/block_scout_web/assets/js/lib/banner.js

// Comment out tất cả code ads
/*
if (showAd()) {
  // ... ads code
} else {
  $('.ad-container').hide()
}
*/

// Chỉ giữ lại
$('.ad-container').hide()
```

#### 3.3. Disable text ads trong `text_ad.js`:
```javascript
// apps/block_scout_web/assets/js/lib/text_ad.js

$(function () {
  // Comment out ads code
  // if (showAd()) {
  //   fetchTextAdData()
  // }
})
```

### Bước 4: Ẩn Elements Quảng Cáo

#### 4.1. Thêm CSS để ẩn ad containers:
```scss
// apps/block_scout_web/assets/css/theme/poatc-theme.scss

// Ẩn tất cả ad containers
.ad-container,
.js-ad-dependant-mb-2,
.js-ad-dependant-mb-3,
.js-ad-dependant-mb-5-reverse,
.ad-banner,
.text-ad {
  display: none !important;
}

// Ẩn các elements liên quan đến ads
[class*="ad-"],
[id*="ad-"],
[class*="banner-"] {
  display: none !important;
}
```

## 🎨 Custom Theme Examples

### Theme 1: Dark Professional
```scss
// Dark theme cho POATC
$primary-color: #00d4ff;
$secondary-color: #1a1a1a;
$background-color: #0f0f0f;
$card-background: #1a1a1a;
$text-color: #ffffff;
$border-color: #333333;
```

### Theme 2: Light Clean
```scss
// Light theme sạch sẽ
$primary-color: #2563eb;
$secondary-color: #f8fafc;
$background-color: #ffffff;
$card-background: #f8fafc;
$text-color: #1e293b;
$border-color: #e2e8f0;
```

### Theme 3: POATC Brand
```scss
// Theme theo brand POATC
$primary-color: #ff6b35;  // Orange
$secondary-color: #2d3748; // Dark gray
$background-color: #f7fafc;
$card-background: #ffffff;
$text-color: #2d3748;
$border-color: #e2e8f0;
```

## 🔧 Advanced Customization

### 1. **Tùy chỉnh Header**
```scss
// apps/block_scout_web/assets/css/theme/poatc-theme.scss

.header {
  background: linear-gradient(135deg, $primary-color, $secondary-color);
  box-shadow: 0 2px 10px rgba(0,0,0,0.1);
  
  .logo {
    max-height: 40px;
    filter: brightness(0) invert(1); // Đổi màu logo thành trắng
  }
}
```

### 2. **Tùy chỉnh Cards**
```scss
.card {
  border-radius: $border-radius;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
  border: 1px solid $border-color;
  background: $card-background;
  
  &:hover {
    box-shadow: 0 4px 16px rgba(0,0,0,0.15);
    transform: translateY(-2px);
    transition: all 0.3s ease;
  }
}
```

### 3. **Tùy chỉnh Buttons**
```scss
.btn-primary {
  background: linear-gradient(135deg, $primary-color, darken($primary-color, 10%));
  border: none;
  border-radius: $button-border-radius;
  font-weight: 600;
  
  &:hover {
    background: linear-gradient(135deg, darken($primary-color, 5%), darken($primary-color, 15%));
    transform: translateY(-1px);
  }
}
```

## 🚀 Build và Deploy

### 1. **Build frontend**
```bash
cd apps/block_scout_web/assets
npm install
npm run build
```

### 2. **Rebuild Docker containers**
```bash
cd docker-compose
docker-compose down
docker-compose up --build -d
```

### 3. **Kiểm tra kết quả**
- Truy cập http://localhost:80
- Kiểm tra logo mới
- Kiểm tra theme mới
- Xác nhận không còn quảng cáo

## 📁 File Structure Summary

```
blockscout-8.0.0/
├── apps/block_scout_web/assets/
│   ├── css/theme/
│   │   ├── poatc-theme.scss     # ← Theme chính
│   │   ├── _variables.scss      # ← Biến màu sắc
│   │   └── _colors.scss         # ← Định nghĩa màu
│   ├── js/lib/
│   │   ├── ad.js               # ← Disable ads
│   │   ├── banner.js           # ← Disable banners
│   │   └── text_ad.js          # ← Disable text ads
│   └── static/images/
│       ├── logo.svg            # ← Logo chính
│       ├── logo-light.svg      # ← Logo dark mode
│       └── favicon.ico         # ← Favicon
└── docker-compose/
    └── docker-compose.yml      # ← Config Docker
```

## ⚠️ Lưu Ý Quan Trọng

### 1. **Backup Files**
- Luôn backup files gốc trước khi chỉnh sửa
- Test trên development trước khi deploy production

### 2. **Performance**
- Optimize images (SVG cho logo, WebP cho photos)
- Minify CSS/JS sau khi customize
- Test trên các thiết bị khác nhau

### 3. **Browser Compatibility**
- Test trên Chrome, Firefox, Safari, Edge
- Kiểm tra responsive design
- Test dark/light mode

### 4. **Updates**
- Khi update Blockscout, backup customizations
- Merge changes từ upstream cẩn thận
- Test thoroughly sau mỗi update

## 🎉 Kết Quả Mong Đợi

Sau khi hoàn thành:
- ✅ Logo POATC hiển thị thay vì logo Blockscout
- ✅ Theme màu sắc phù hợp với brand
- ✅ Không còn quảng cáo nào
- ✅ Giao diện đẹp và professional
- ✅ Responsive trên mọi thiết bị

---

**🎨 Chúc bạn có giao diện Blockscout đẹp và phù hợp với POATC!**
