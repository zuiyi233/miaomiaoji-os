---
name: ai-human-writing
description: This skill transforms AI-generated text into authentic, human-like writing by eliminating AI detection markers through entropy balancing, cognitive simulation, and natural expression techniques. Use this skill when the user wants to write content that avoids AI detection, needs to add personality and voice to their writing, or wants to make their text sound more natural, emotional, and authentically human.
---

# AI真人感写作技能 (AI Human-like Writing Skill)

## 概述 (Overview)

将AI生成的文本转化为具有真人质感的自然写作。通过熵值平衡、认知模拟、缺陷植入等技术，消除AI特征，注入人类的不完美与个性。

## 使用场景 (When to Use)

- 需要避免AI检测的写作任务
- 为内容注入个性化表达和情感
- 使文本更具真实感和可读性
- 创作具有独特"人声"的叙事内容

## 核心工作流程 (Core Workflow)

### Step 1: 内容分析 (Content Analysis)

在修改前，分析文本的AI特征：

1. **词汇层面检查**
   - 是否使用AI高频词（因此、然而、诚然、显然）
   - 是否存在过度学术化/文言化的表达
   - 同义词替换是否过于频繁

2. **句法层面检查**
   - 句子长度是否过于均匀
   - 是否全是完整的主谓宾结构
   - 标点符号使用是否过于规范

3. **段落层面检查**
   - 信息密度是否均匀
   - 逻辑是否过于严密
   - 是否存在思维跳跃

### Step 2: 熵值平衡 (Entropy Balancing)

实施"情绪过山车"技术：

- **高熵区域**（放飞自我）：
  - 怼人、吐槽、情绪爆发时彻底放飞
  - 使用短句、语气词、感叹号
  - 插入怪异的比喻和骚话

- **低熵区域**（稍作收敛）：
  - 背景描述、日常对话时相对正常
  - 但要为随时"发疯"做好铺垫

### Step 3: 词汇指纹植入 (Lexical Fingerprint)

参考 `references/personal-voice.md` 中的语料库：

1. **高频词替换**
   - "因此" → "所以"
   - "然而" → "但是/可是/不过"
   - "值得注意的是" → 直接说事

2. **口头禅植入**
   - 根据场景插入个人标志性表达
   - 保持一定频率，但避免过度

3. **禁用词清理**
   - 删除"殊不知"、"众所周知"等装逼词汇

### Step 4: 句式重构 (Sentence Restructuring)

1. **打破均匀节奏**
   - 长短句交错
   - 紧张时用短句（"完了。""我靠。"）
   - 叙述时用长短结合

2. **引入不完整句**
   - 省略主语（"就离谱。"）
   - 话说到一半（"你要是再……算了。"）

3. **增加口语化特征**
   - 使用"那什么"、"怎么说呢"、"哎（↗）"

### Step 5: 缺陷植入 (Flaw Injection)

根据文本长度植入可控缺陷：

- **>2000字**: 偶尔来句语序颠倒或不合语法的句子
- **>1000字**: 故意写个别字（"的地得"不分）
- **>500字**: 标点符号跟着心情走（!!! / ......）

### Step 6: 认知模拟 (Cognitive Simulation)

1. **记忆缺陷**
   - 引入不确定性（"我记得……上周？不对，好像是上个月"）

2. **注意力偏差**
   - 只抓自己想说的点猛攻
   - 其他"显而易见"的东西无所谓

3. **确认偏见**
   - 内心独白充满偏见（"我看他那样，就知道没憋什么好屁"）

### Step 7: 质量验证 (Quality Validation)

使用 `references/checklist.md` 进行最终检查：

- [ ] 是不是每句话都太正经了？（必须不是）
- [ ] 是不是有连续两段话读起来很顺畅？（最好不要）
- [ ] 是不是所有句子都主谓宾齐全？（绝对不行）
- [ ] 是不是一个错别字都没有？（赶紧加一个）
- [ ] 是不是用词太高级了？（换成大白话）
- [ ] 是不是逻辑太清晰了？（把它搅浑）

## 快速参考 (Quick Reference)

### 常用替换表

| AI特征词 | 自然替代 |
|---------|---------|
| 因此 | 所以、这不就得了 |
| 然而 | 但是、可是、不过 |
| 值得注意的是 | 直接说事 |
| 首先/其次/最后 | 删掉，直接说 |
| 突然 | 猛地 |
| 立刻/马上 | 立马 |
| 难道 | 难不成 |

### 节奏控制

- **紧张**: 短句！断句！感叹号！
- **吐槽**: 反问句 + 网络梗
- **叙述**: 长短句结合，随时准备发疯
- **内心OS**: 思维跳跃，前后不搭

## 参考资料 (References)

- `references/anti-detection-guide.md` - 详细的反检测策略指南
- `references/personal-voice.md` - 个人语料库与表达指纹
- `references/checklist.md` - 快速检查清单

## 注意事项 (Important Notes)

1. **适度原则**: 别疯得太过，让人看不懂就行
2. **场景适配**: 根据内容类型调整"发疯"程度
3. **保持一致**: 同一文本中保持语气和风格的一致性
4. **保留精华**: 不要为了"像人"而牺牲内容质量
