# Spec: ui-command-palette

**Status**: Draft
**Created**: 2026-03-02

---

## Overview

本规范定义了 OpsPilot PaaS 平台的命令面板功能。命令面板提供快速导航、搜索和操作的能力，通过键盘快捷键 (Cmd+K / Ctrl+K) 唤起，提升用户操作效率。

---

## Requirements

### REQ-1: 命令面板触发

系统 SHALL 提供全局命令面板，通过键盘快捷键唤起。

**快捷键**:
- macOS SHALL 使用 Cmd+K
- Windows/Linux SHALL 使用 Ctrl+K
- 快捷键 SHALL 在任何页面生效
- 再次按下快捷键 SHALL 关闭命令面板
- 按 Esc SHALL 关闭命令面板

#### Scenario: 打开命令面板

**WHEN** 用户按下 Cmd+K (macOS) 或 Ctrl+K (Windows/Linux)
**THEN** 命令面板 SHALL 打开
**AND** 搜索框 SHALL 自动获得焦点
**AND** 背景遮罩 SHALL 显示

---

### REQ-2: 命令面板样式

命令面板 SHALL 采用现代化的对话框样式。

**对话框样式**:
- 宽度 SHALL 为 640px
- 最大高度 SHALL 为 480px
- 背景色 SHALL 为白色
- 圆角 SHALL 为 12px
- 阴影 SHALL 为 2xl
- Position SHALL 为 fixed，居中显示
- Z-Index SHALL 为 1060

**背景遮罩**:
- 背景色 SHALL 为 rgba(0, 0, 0, 0.45)
- SHALL 有 4px 的模糊效果
- 点击遮罩 SHALL 关闭命令面板

**动画**:
- 进入动画 SHALL 为 200ms
- 进入时 SHALL 从 95% 缩放到 100%
- 进入时 SHALL 从透明度 0 到 1
- 退出动画 SHALL 为 150ms

#### Scenario: 命令面板样式

**WHEN** 命令面板打开
**THEN** SHALL 显示在屏幕中央
**AND** 背景 SHALL 有模糊效果
**AND** 对话框 SHALL 有缩放动画

---

### REQ-3: 搜索框

命令面板 SHALL 包含搜索框，用于输入命令或搜索。

**搜索框样式**:
- 高度 SHALL 为 56px
- 内边距 SHALL 为 0 20px
- 字体 SHALL 为 16px
- 边框 SHALL 为 none
- 底部边框 SHALL 为 1px solid #e9ecef
- Placeholder SHALL 为 "搜索或执行命令..."
- 搜索图标 SHALL 显示在左侧

**搜索行为**:
- SHALL 支持实时搜索（无去抖）
- SHALL 支持模糊匹配
- SHALL 高亮匹配的文字

#### Scenario: 搜索命令

**WHEN** 用户输入 "service"
**THEN** 命令列表 SHALL 过滤显示包含 "service" 的命令
**AND** 匹配的文字 SHALL 高亮显示
**AND** 第一个结果 SHALL 自动选中

---

### REQ-4: 命令分组

命令面板 SHALL 将命令按类型分组显示。

**分组**:
- 导航 (Navigation)
- 搜索 (Search)
- 操作 (Actions)
- 最近使用 (Recent)

**分组样式**:
- 分组标题字体 SHALL 为 12px, Semibold (600)
- 分组标题颜色 SHALL 为 Gray 500
- 分组标题 SHALL 大写，字间距 0.5px
- 分组标题内边距 SHALL 为 8px 20px
- 分组之间 SHALL 有 4px 间距

#### Scenario: 命令分组显示

**WHEN** 用户打开命令面板且搜索框为空
**THEN** SHALL 显示"最近使用"分组
**AND** SHALL 显示"导航"分组
**AND** SHALL 显示"操作"分组
**AND** 每个分组 SHALL 有标题

---

### REQ-5: 命令项

命令面板 SHALL 显示命令项列表。

**命令项样式**:
- 高度 SHALL 为 48px
- 内边距 SHALL 为 0 20px
- 字体 SHALL 为 14px
- 默认背景色 SHALL 为透明
- 选中背景色 SHALL 为 Gray 100
- 悬停背景色 SHALL 为 Gray 100
- 过渡动画 SHALL 为 100ms

**命令项内容**:
- 左侧 SHALL 显示图标（18px）
- 图标与文字间距 SHALL 为 12px
- 文字颜色 SHALL 为 Gray 900
- 右侧 SHALL 显示快捷键提示（如果有）
- 快捷键字体 SHALL 为 12px
- 快捷键颜色 SHALL 为 Gray 500
- 快捷键 SHALL 使用 Kbd 样式

**Kbd 样式**:
- 背景色 SHALL 为 Gray 100
- 边框 SHALL 为 1px solid #dee2e6
- 圆角 SHALL 为 4px
- 内边距 SHALL 为 2px 6px
- 字体 SHALL 为等宽字体

#### Scenario: 命令项显示

**WHEN** 用户查看命令列表
**THEN** 每个命令 SHALL 显示图标和文字
**AND** 导航命令 SHALL 显示目标页面名称
**AND** 操作命令 SHALL 显示操作名称
**AND** 有快捷键的命令 SHALL 在右侧显示快捷键

---

### REQ-6: 键盘导航

命令面板 SHALL 支持完整的键盘导航。

**键盘操作**:
- ↑ (Up) SHALL 选择上一个命令
- ↓ (Down) SHALL 选择下一个命令
- Enter SHALL 执行选中的命令
- Esc SHALL 关闭命令面板
- Tab SHALL 在分组间切换（可选）

**选中状态**:
- 选中的命令 SHALL 高亮显示
- 选中的命令 SHALL 滚动到可见区域
- 默认 SHALL 选中第一个命令

#### Scenario: 键盘导航

**WHEN** 用户按下 ↓ 键
**THEN** 选中的命令 SHALL 移动到下一个
**AND** 新选中的命令 SHALL 高亮显示
**AND** 如果需要，SHALL 滚动到可见区域

**WHEN** 用户按下 Enter
**THEN** SHALL 执行选中的命令
**AND** 命令面板 SHALL 关闭

---

### REQ-7: 导航命令

命令面板 SHALL 提供快速导航命令。

**导航命令列表**:
- 主控台 (Dashboard)
- 服务列表 (Services)
- 服务详情 (Service Detail) - 如果当前在服务相关页面
- 部署管理 (Deployment)
- 主机管理 (Hosts)
- 监控告警 (Monitoring)
- 配置中心 (Config)
- 任务调度 (Jobs)
- 设置 (Settings)

**执行行为**:
- 执行导航命令 SHALL 跳转到对应页面
- 跳转后 SHALL 关闭命令面板

#### Scenario: 导航到服务列表

**WHEN** 用户在命令面板中选择"服务列表"并按 Enter
**THEN** SHALL 跳转到服务列表页
**AND** 命令面板 SHALL 关闭

---

### REQ-8: 搜索命令

命令面板 SHALL 提供全局搜索功能。

**搜索类型**:
- 搜索服务
- 搜索主机
- 搜索部署记录
- 搜索配置项

**搜索行为**:
- 输入 "service:" 或 "服务:" SHALL 触发服务搜索
- 输入 "host:" 或 "主机:" SHALL 触发主机搜索
- 输入 "deploy:" 或 "部署:" SHALL 触发部署搜索
- 搜索结果 SHALL 实时显示
- 选中搜索结果 SHALL 跳转到详情页

#### Scenario: 搜索服务

**WHEN** 用户输入 "service:api"
**THEN** SHALL 显示名称包含 "api" 的服务
**AND** 每个结果 SHALL 显示服务名称和状态
**AND** 选中结果并按 Enter SHALL 跳转到服务详情页

---

### REQ-9: 操作命令

命令面板 SHALL 提供快速操作命令。

**操作命令列表**:
- 创建新服务
- 部署服务
- 添加主机
- 创建配置
- 创建任务

**执行行为**:
- 执行操作命令 SHALL 打开对应的创建/编辑页面或对话框
- 命令面板 SHALL 关闭

#### Scenario: 创建新服务

**WHEN** 用户在命令面板中选择"创建新服务"并按 Enter
**THEN** SHALL 跳转到服务创建页面
**AND** 命令面板 SHALL 关闭

---

### REQ-10: 最近使用

命令面板 SHALL 记录和显示最近使用的命令。

**记录规则**:
- SHALL 记录最近 5 个使用的命令
- SHALL 记录命令类型和参数
- SHALL 持久化到 localStorage

**显示规则**:
- 最近使用 SHALL 显示在顶部
- 最近使用 SHALL 有独立的分组
- 如果搜索框为空，SHALL 默认显示最近使用

#### Scenario: 最近使用命令

**WHEN** 用户打开命令面板且搜索框为空
**THEN** SHALL 显示"最近使用"分组
**AND** SHALL 显示最近 5 个使用的命令
**AND** 命令 SHALL 按使用时间倒序排列

---

### REQ-11: 空状态

命令面板 SHALL 在无结果时显示空状态。

**空状态样式**:
- SHALL 显示在列表区域中央
- 图标尺寸 SHALL 为 48px
- 图标颜色 SHALL 为 Gray 300
- 提示文字字体 SHALL 为 14px
- 提示文字颜色 SHALL 为 Gray 500
- 提示文字 SHALL 为 "未找到匹配的命令"

#### Scenario: 搜索无结果

**WHEN** 用户输入的搜索词没有匹配结果
**THEN** SHALL 显示空状态
**AND** SHALL 显示 "未找到匹配的命令" 提示

---

### REQ-12: 响应式适配

命令面板 SHALL 支持响应式适配。

**移动端 (< 768px)**:
- 宽度 SHALL 为 90vw
- 最大宽度 SHALL 为 480px
- 高度 SHALL 适应内容
- 最大高度 SHALL 为 70vh

#### Scenario: 移动端命令面板

**WHEN** 用户在移动设备上打开命令面板
**THEN** 宽度 SHALL 适应屏幕
**AND** 高度 SHALL 适应内容
**AND** 所有功能 SHALL 正常工作

---

## Implementation Notes

### 使用 cmdk 库

```typescript
// CommandPalette.tsx
import { Command } from 'cmdk';
import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';

export const CommandPalette = () => {
  const [open, setOpen] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === 'k' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setOpen((open) => !open);
      }
    };

    document.addEventListener('keydown', down);
    return () => document.removeEventListener('keydown', down);
  }, []);

  return (
    <Command.Dialog open={open} onOpenChange={setOpen}>
      <Command.Input placeholder="搜索或执行命令..." />
      <Command.List>
        <Command.Empty>未找到匹配的命令</Command.Empty>

        <Command.Group heading="导航">
          <Command.Item onSelect={() => navigate('/')}>
            <HomeOutlined />
            <span>主控台</span>
          </Command.Item>
          <Command.Item onSelect={() => navigate('/services')}>
            <AppstoreOutlined />
            <span>服务列表</span>
          </Command.Item>
        </Command.Group>

        <Command.Group heading="操作">
          <Command.Item onSelect={() => navigate('/services/new')}>
            <PlusOutlined />
            <span>创建新服务</span>
          </Command.Item>
        </Command.Group>
      </Command.List>
    </Command.Dialog>
  );
};
```

### 最近使用记录

```typescript
// useRecentCommands.ts
const RECENT_COMMANDS_KEY = 'recent_commands';

export const useRecentCommands = () => {
  const [recent, setRecent] = useState<Command[]>([]);

  useEffect(() => {
    const stored = localStorage.getItem(RECENT_COMMANDS_KEY);
    if (stored) {
      setRecent(JSON.parse(stored));
    }
  }, []);

  const addRecent = (command: Command) => {
    const updated = [command, ...recent.filter(c => c.id !== command.id)].slice(0, 5);
    setRecent(updated);
    localStorage.setItem(RECENT_COMMANDS_KEY, JSON.stringify(updated));
  };

  return { recent, addRecent };
};
```

---

## References

- [cmdk](https://cmdk.paco.me/)
- [Command Palette Best Practices](https://www.commandpalette.com/)

