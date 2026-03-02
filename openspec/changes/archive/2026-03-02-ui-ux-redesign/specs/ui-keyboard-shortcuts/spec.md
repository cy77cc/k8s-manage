# Spec: ui-keyboard-shortcuts

**Status**: Draft
**Created**: 2026-03-02

---

## Overview

本规范定义了 OpsPilot PaaS 平台的键盘快捷键系统。键盘快捷键提供快速导航和操作的能力，提升高级用户的操作效率。

---

## Requirements

### REQ-1: 全局快捷键

系统 SHALL 提供全局快捷键，在任何页面生效。

**全局快捷键列表**:
- `Cmd/Ctrl + K`: 打开命令面板
- `/`: 聚焦搜索框
- `Esc`: 关闭模态框/面板/抽屉
- `?`: 显示快捷键帮助

**实现规则**:
- 快捷键 SHALL 在输入框获得焦点时禁用（除 Esc）
- 快捷键 SHALL 不与浏览器默认快捷键冲突
- macOS SHALL 使用 Cmd 键
- Windows/Linux SHALL 使用 Ctrl 键

#### Scenario: 聚焦搜索框

**WHEN** 用户按下 `/` 键
**THEN** 页面搜索框 SHALL 获得焦点
**AND** 如果页面没有搜索框，SHALL 打开命令面板

---

### REQ-2: 导航快捷键

系统 SHALL 提供导航快捷键，用于快速跳转页面。

**导航快捷键列表**:
- `g + h`: 跳转到主控台 (Home)
- `g + s`: 跳转到服务列表 (Services)
- `g + d`: 跳转到部署管理 (Deployment)
- `g + m`: 跳转到监控告警 (Monitoring)
- `g + o`: 跳转到主机管理 (hOsts)
- `g + c`: 跳转到配置中心 (Config)
- `g + j`: 跳转到任务调度 (Jobs)

**实现规则**:
- SHALL 使用组合键模式（先按 g，再按目标键）
- 按下 g 后 SHALL 有 1 秒的等待时间
- 超时或按下其他键 SHALL 取消组合
- 组合键过程中 SHALL 显示提示（可选）

#### Scenario: 跳转到服务列表

**WHEN** 用户按下 `g` 然后按下 `s`
**THEN** SHALL 跳转到服务列表页
**AND** 如果已在服务列表页，SHALL 滚动到顶部

---

### REQ-3: 列表快捷键

系统 SHALL 在列表页面提供快捷键。

**列表快捷键**:
- `j`: 选择下一项
- `k`: 选择上一项
- `Enter`: 打开选中项
- `Space`: 选中/取消选中（多选模式）
- `Cmd/Ctrl + A`: 全选
- `Cmd/Ctrl + D`: 取消全选

**实现规则**:
- 列表项 SHALL 有选中状态（高亮显示）
- 选中的项 SHALL 滚动到可见区域
- 默认 SHALL 选中第一项

#### Scenario: 列表导航

**WHEN** 用户在服务列表页按下 `j`
**THEN** 选中的服务 SHALL 移动到下一个
**AND** 新选中的服务 SHALL 高亮显示
**AND** 如果需要，SHALL 滚动到可见区域

**WHEN** 用户按下 `Enter`
**THEN** SHALL 打开选中的服务详情页

---

### REQ-4: 表格快捷键

系统 SHALL 在表格中提供快捷键。

**表格快捷键**:
- `j`: 选择下一行
- `k`: 选择上一行
- `h`: 向左滚动
- `l`: 向右滚动
- `Enter`: 打开选中行
- `Space`: 选中/取消选中行

#### Scenario: 表格导航

**WHEN** 用户在表格中按下 `j`
**THEN** 选中的行 SHALL 移动到下一行
**AND** 新选中的行 SHALL 高亮显示

---

### REQ-5: 模态框快捷键

系统 SHALL 在模态框中提供快捷键。

**模态框快捷键**:
- `Esc`: 关闭模态框
- `Enter`: 确认操作（如果有确认按钮）
- `Tab`: 在表单字段间切换
- `Shift + Tab`: 反向切换

**实现规则**:
- Enter 快捷键 SHALL 仅在非文本域时生效
- Esc SHALL 在任何时候生效
- 危险操作 SHALL 需要二次确认，不能直接用 Enter

#### Scenario: 关闭模态框

**WHEN** 用户在模态框中按下 `Esc`
**THEN** 模态框 SHALL 关闭
**AND** 如果有未保存的更改，SHALL 显示确认对话框

---

### REQ-6: 编辑器快捷键

系统 SHALL 在代码/配置编辑器中提供快捷键。

**编辑器快捷键**:
- `Cmd/Ctrl + S`: 保存
- `Cmd/Ctrl + F`: 查找
- `Cmd/Ctrl + H`: 替换
- `Cmd/Ctrl + Z`: 撤销
- `Cmd/Ctrl + Shift + Z`: 重做
- `Cmd/Ctrl + /`: 注释/取消注释

**实现规则**:
- 编辑器快捷键 SHALL 优先于全局快捷键
- SHALL 使用 Monaco Editor 或 CodeMirror 的内置快捷键

#### Scenario: 保存配置

**WHEN** 用户在配置编辑器中按下 `Cmd/Ctrl + S`
**THEN** SHALL 保存配置
**AND** SHALL 显示保存成功提示
**AND** 不应触发浏览器的保存页面功能

---

### REQ-7: 快捷键帮助

系统 SHALL 提供快捷键帮助对话框。

**触发方式**:
- 按下 `?` 键
- 点击页面右下角的帮助按钮

**帮助对话框样式**:
- 宽度 SHALL 为 640px
- 最大高度 SHALL 为 80vh
- 背景色 SHALL 为白色
- 圆角 SHALL 为 12px
- 阴影 SHALL 为 xl

**内容组织**:
- SHALL 按类别分组（全局、导航、列表、编辑器等）
- 每个快捷键 SHALL 显示：按键 + 描述
- 按键 SHALL 使用 Kbd 样式
- SHALL 支持搜索快捷键

#### Scenario: 查看快捷键帮助

**WHEN** 用户按下 `?`
**THEN** SHALL 显示快捷键帮助对话框
**AND** SHALL 显示所有可用的快捷键
**AND** 快捷键 SHALL 按类别分组
**AND** 按下 `Esc` SHALL 关闭帮助对话框

---

### REQ-8: 快捷键冲突处理

系统 SHALL 正确处理快捷键冲突。

**优先级规则**:
1. 输入框内的快捷键（最高优先级）
2. 编辑器快捷键
3. 页面级快捷键
4. 全局快捷键（最低优先级）

**冲突处理**:
- 输入框获得焦点时，SHALL 禁用大部分全局快捷键
- 编辑器获得焦点时，SHALL 禁用冲突的全局快捷键
- Esc 键 SHALL 始终生效

#### Scenario: 输入框中的快捷键

**WHEN** 用户在输入框中输入 "/"
**THEN** SHALL 输入 "/" 字符
**AND** 不应触发聚焦搜索框的快捷键

---

### REQ-9: 快捷键提示

系统 SHALL 在适当的位置显示快捷键提示。

**提示位置**:
- 按钮 Tooltip 中显示快捷键
- 菜单项右侧显示快捷键
- 命令面板中显示快捷键

**提示样式**:
- SHALL 使用 Kbd 样式
- 字体 SHALL 为 12px
- 颜色 SHALL 为 Gray 500

#### Scenario: 按钮快捷键提示

**WHEN** 用户悬停在"创建服务"按钮上
**THEN** Tooltip SHALL 显示 "创建服务 (Cmd+N)"
**AND** 快捷键 SHALL 使用 Kbd 样式

---

### REQ-10: 自定义快捷键

系统 SHALL 允许用户自定义快捷键（可选功能）。

**自定义规则**:
- SHALL 在设置页面提供快捷键配置
- SHALL 检测快捷键冲突
- SHALL 提供重置为默认的选项
- SHALL 持久化到用户配置

#### Scenario: 自定义快捷键

**WHEN** 用户在设置中修改"打开命令面板"的快捷键为 `Cmd+P`
**THEN** SHALL 保存配置
**AND** 新快捷键 SHALL 立即生效
**AND** 如果与其他快捷键冲突，SHALL 显示警告

---

### REQ-11: 快捷键禁用

系统 SHALL 允许在特定场景禁用快捷键。

**禁用场景**:
- 输入框获得焦点时
- 编辑器获得焦点时
- 模态框中的输入框获得焦点时

**实现规则**:
- SHALL 检测焦点元素类型
- SHALL 根据元素类型决定是否禁用快捷键
- Esc 键 SHALL 始终生效

#### Scenario: 输入框中禁用快捷键

**WHEN** 用户在搜索框中输入内容
**THEN** 导航快捷键 SHALL 被禁用
**AND** 用户可以正常输入字符
**AND** 按下 `Esc` SHALL 清空搜索框或失去焦点

---

### REQ-12: 响应式适配

快捷键系统 SHALL 支持移动端。

**移动端规则**:
- 移动端 SHALL 禁用大部分键盘快捷键
- SHALL 保留 Esc 键功能（通过返回按钮）
- 快捷键帮助 SHALL 适配小屏幕

#### Scenario: 移动端快捷键

**WHEN** 用户在移动设备上使用系统
**THEN** 键盘快捷键 SHALL 被禁用
**AND** 返回按钮 SHALL 等同于 Esc 键
**AND** 快捷键帮助 SHALL 不显示（或显示简化版本）

---

## Implementation Notes

### 快捷键监听

```typescript
// useKeyboardShortcuts.ts
import { useEffect } from 'react';

export const useKeyboardShortcuts = () => {
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // 检查是否在输入框中
      const target = e.target as HTMLElement;
      const isInput = ['INPUT', 'TEXTAREA', 'SELECT'].includes(target.tagName);
      const isContentEditable = target.isContentEditable;

      if (isInput || isContentEditable) {
        // 仅允许 Esc 键
        if (e.key === 'Escape') {
          target.blur();
        }
        return;
      }

      // 全局快捷键
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        // 打开命令面板
      }

      if (e.key === '/') {
        e.preventDefault();
        // 聚焦搜索框
      }

      if (e.key === '?') {
        e.preventDefault();
        // 显示快捷键帮助
      }

      // 导航快捷键 (g + x)
      // ...
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, []);
};
```

### Kbd 组件

```typescript
// Kbd.tsx
export const Kbd: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  return (
    <kbd className="kbd">
      {children}
    </kbd>
  );
};

// kbd.css
.kbd {
  display: inline-block;
  padding: 2px 6px;
  font-family: 'SF Mono', 'Monaco', monospace;
  font-size: 12px;
  line-height: 1;
  color: #495057;
  background-color: #f8f9fa;
  border: 1px solid #dee2e6;
  border-radius: 4px;
  box-shadow: 0 1px 0 rgba(0, 0, 0, 0.05);
}
```

---

## References

- [Keyboard Shortcuts Best Practices](https://www.nngroup.com/articles/keyboard-shortcuts/)
- [Web Accessibility - Keyboard](https://www.w3.org/WAI/WCAG21/Understanding/keyboard.html)

