---
layout: post
title: "Prompt Engineering 课程笔记"
date:   2025-11-23
tags: [AI]
comments: true
author: xiaodp
toc: true
---
本文记录Prompt Engineering课程学习记录

<!-- more -->

# Prompt Engineering 课程笔记

https://www.bilibili.com/video/BV1173jzNELG/?spm_id_from=333.1391.0.0&vd_source=7fdf8ec1d7ea5aeaf56c60a03a23c1a7

## 核心原则（Principles of Prompting）

### Principle 1: Write clear and specific instructions（编写清晰且具体的指令）

#### Tactic 1: Use delimiters（使用分隔符）

使用分隔符来清晰地标识输入的不同部分，避免提示词注入（prompt injection）。

**常用的分隔符类型：**
- 三重引号：`"""`
- 三重反引号：``` ```
- 三重破折号：`---`
- 尖括号：`< >`
- XML 标签：`<tag> </tag>`

#### Tactic 2: Ask for structured output（要求结构化输出）

要求模型以特定格式输出，便于后续处理。

**常用格式：**
- HTML
- JSON

#### Tactic 3: Check whether conditions are satisfied（检查是否满足条件）

在执行任务前，检查完成任务所需的假设条件是否满足。

#### Tactic 4: Few-shot prompting（少样本提示）

提供成功完成任务的示例，然后要求模型执行相同的任务。

**示例：保持一致的风格**

通过提供一个示例，让模型学习并保持特定的风格（如"grandparent"风格）：

**提示词：**
```
Your task is to answer in a consistent style.

<child>: Teach me about patience.
<grandparent>: The river that carves the deepest 
valley flows from a modest spring; the 
grandest symphony originates from a single note; 
the most intricate tapestry begins with a solitary thread.

<child>: Teach me about resilience.
```

**模型输出：**
```
<grandparent>: Resilience is like a tree that bends with the wind but never breaks. 
It is the ability to bounce back from adversity and keep moving forward, even when 
things get tough. Just like a tree that grows stronger with each storm it weathers, 
resilience is a quality that can be developed and strengthened over time.
```

**说明：** 这个示例展示了 few-shot prompting 如何引导模型采用特定的语调、风格或角色来回答问题。

#### 综合示例：结合多个策略

以下示例展示了如何在一个提示词中结合多个策略（使用分隔符、结构化输出、清晰的步骤指令）：

**输入文本：**
```python
text = f"""
In a charming village, siblings Jack and Jill set out on a quest to fetch water from a hilltop well. 
As they climbed, singing joyfully, misfortune struck-Jack tripped on a stone and tumbled down the hill, 
with Jill following suit. Though slightly battered, the pair returned home to comforting embraces. 
Despite the mishap, their adventurous spirits remained undimmed, and they continued exploring with delight.
"""
```

**提示词（Example 1）：**
```python
prompt_1 = f"""
Perform the following actions:
1. Summarize the following text delimited by triple backticks with 1 sentence.
2. Translate the summary into French.
3. List each name in the French summary.
4. Output a json object that contains the following keys: french_summary, num_names.

Separate your answers with line breaks.

Text: ```{text}```
"""
```

**说明：** 这个示例同时运用了：
- **分隔符**：使用三重反引号（```）来标识输入文本
- **结构化输出**：要求输出 JSON 格式
- **清晰的步骤**：使用编号列表明确指定任务步骤
- **格式要求**：要求用换行符分隔答案

**提示词（Example 2）：使用尖括号分隔符和详细输出格式**

```python
prompt_2 = f"""
Perform the following actions:
1. Summarize the following text delimited by < > into a single sentence.
2. Translate the summary into French.
3. List each name found in the French summary.
4. Output a JSON object that contains the following keys: french_summary, num_names.

Use the following format:
Text: <text to summarize>
Summary: <summary>
Translation: <summary translation>
Names: <list of names in French summary>
Output JSON: <json with summary and num_names>

Text: <{text}>
"""
```

**模型输出：**
```
Summary: Jack and Jill go on a quest to fetch water, but misfortune strikes and they tumble down the hill, returning home slightly battered but with their adventurous spirits undimmed.

Translation: Jack et Jill partent en quête d'eau, mais un malheur frappe et ils tombent de la colline, rentrant chez eux légèrement meurtris mais avec leurs esprits aventureux intacts.

Names: Jack, Jill

Output JSON: {"french_summary": "Jack et Jill partent en quête d'eau, mais un malheur frappe et ils tombent de la colline, rentrant chez eux légèrement meurtris mais avec leurs esprits aventureux intacts.", "num_names": 2}
```

**说明：** Example 2 展示了：
- **不同的分隔符**：使用尖括号 `< >` 代替三重反引号
- **更详细的输出格式**：明确指定每个输出部分的格式和标签
- **结构化输出**：要求 JSON 格式，并指定了具体的键名
- **模型的能力**：即使格式说明中有小错误（如提到"Italian summary"），模型也能根据明确的指令正确执行（提取法语摘要中的名字）

**对比两个示例：**
- Example 1 更简洁，使用换行符分隔答案
- Example 2 更详细，使用标签明确标识每个输出部分，便于后续解析

---

## Principle 2: Give the model time to think（给模型时间思考）

### Tactic 1: Specify the steps to complete a task（指定完成任务的步骤）

将复杂任务分解为明确的步骤，让模型按顺序执行：

```
Step 1: ...
Step 2: ...
...
Step N: ...
```

通过明确指定步骤，可以帮助模型更好地组织思路，避免遗漏关键环节。

### Tactic 2: Instruct the model to work out its own solution before rushing to a conclusion（指导模型在匆忙得出结论之前先自己解决问题）

要求模型先进行推理和思考，然后再给出最终答案。这样可以：
- 让模型有更多时间进行深入思考
- 减少直接给出错误答案的可能性
- 提高答案的准确性和可靠性

#### 示例：检查学生解答的正确性

**问题场景：**
判断学生的解答是否正确，但不要直接给出结论，而是先让模型自己解决问题。

**提示词：**
```python
prompt = f"""
Determine if the student's solution is correct or not.

Question:
I'm building a solar power installation and I need help working out the financials.
- Land costs $100 / square foot
- I can buy solar panels for $250 / square foot
- I have a maintenance contract that costs a flat $100k per year, and an additional $10 / square foot
What is the total cost for the first year of operations as a function of the number of square feet.

Student's Solution:
Let x be the size of the installation in square feet.
Costs:
1. Land cost: 100x
2. Solar panel cost: 250x
3. Maintenance cost: 100,000 + 100x
Total cost: 100x + 250x + 100,000 + 100x = 450x + 100,000
"""
```

**说明：** 这个提示词直接要求模型判断学生解答是否正确，但没有要求模型先自己解决问题。更好的做法是要求模型先自己计算，然后再比较判断。

**改进版本的提示词：**

```python
prompt = f"""
Your task is to determine if the student's solution is correct or not.
To solve the problem do the following:
- First, work out your own solution to the problem.
- Then compare your solution to the student's solution and evaluate if the student's solution is correct or not.
Don't decide if the student's solution is correct until you have done the problem yourself.

Use the following format:
Question:
question here
Student's solution:
student's solution here
Actual solution:
steps to work out the solution and your solution here
Is the student's solution the same as actual solution just calculated:
yes or no
Student grade:
correct or incorrect

Question:
I'm building a solar power installation and I need help working out the financials.
- Land costs $100 / square foot
- I can buy solar panels for $250 / square foot
- I have a maintenance contract that costs a flat $100k per year, and an additional $10 / square foot
What is the total cost for the first year of operations as a function of the number of square feet.

Student's Solution:
Let x be the size of the installation in square feet.
Costs:
1. Land cost: 100x
2. Solar panel cost: 250x
3. Maintenance cost: 100,000 + 100x
Total cost: 100x + 250x + 100,000 + 100x = 450x + 100,000
"""
```

**模型输出：**
```
Question:
I'm building a solar power installation and I need help working out the financials.
- Land costs $100 / square foot
- I can buy solar panels for $250 / square foot
- I have a maintenance contract that costs a flat $100k per year, and an additional $10 / square foot
What is the total cost for the first year of operations as a function of the number of square feet.

Student's solution:
Let x be the size of the installation in square feet.
Costs:
1. Land cost: 100x
2. Solar panel cost: 250x
3. Maintenance cost: 100,000 + 100x
Total cost: 100x + 250x + 100,000 + 100x = 450x + 100,000

Actual solution:
Let x be the size of the installation in square feet.
Costs:
1. Land cost: 100x
2. Solar panel cost: 250x
3. Maintenance cost: 100,000 + 10x
Total cost: 100x + 250x + 100,000 + 10x = 360x + 100,000

Is the student's solution the same as actual solution just calculated:
No

Student grade:
Incorrect
```

**关键改进点：**
1. **明确要求模型先自己解决问题**：通过"First, work out your own solution"指令，强制模型先进行计算
2. **结构化输出格式**：指定了详细的输出格式，包括"Actual solution"部分，让模型展示自己的计算过程
3. **避免匆忙判断**：通过"Don't decide if the student's solution is correct until you have done the problem yourself"明确要求模型先完成自己的计算
4. **正确识别错误**：模型正确识别出学生将维护成本误写为 `100x`（应该是 `10x`），因此总成本计算错误（学生：450x + 100,000，正确答案：360x + 100,000）

**对比：**
- **原版本**：模型可能直接判断，容易出错
- **改进版本**：模型先自己计算，然后比较，能够准确识别学生的错误

---

## 模型限制（Model Limitations）

### Hallucination（幻觉）

**定义：** 模型会生成听起来合理但实际上不真实的陈述。

**说明：** 这是语言模型的一个重要限制。模型可能会：
- 编造看似真实但实际不存在的信息
- 提供听起来权威但实际上是错误的内容
- 在不确定的情况下仍然给出肯定的答案

**应对策略：**
- 要求模型从提供的文档中查找信息，而不是依赖其训练数据
- 让模型先引用来源，然后再给出答案
- 对于重要信息，要求模型标注不确定的部分
- 使用分隔符明确指定信息来源，减少幻觉的可能性

#### 示例：模型幻觉问题

**提示词：**
```python
prompt = f"""
Tell me about AeroGlide UltraSlim Smart Toothbrush by Boie
"""
response = get_completion(prompt)
print(response)
```

**模型输出：**
```
The AeroGlide UltraSlim Smart Toothbrush by Boie is a high-tech dental device 
that uses advanced sonic technology to provide a deep and thorough clean. The 
toothbrush features a slim and sleek design that makes it easy to handle and 
maneuver, and it comes with a range of smart features that help optimize your 
brushing routine.

One of the key features of the AeroGlide UltraSlim Smart Toothbrush is its 
advanced sonic technology. This technology uses high-frequency vibrations to 
break up plaque and bacteria, effectively removing tough stains and buildup, 
leaving your teeth clean and refreshed.

In addition to its sonic technology, the AeroGlide UltraSlim Smart Toothbrush 
also comes with a range of smart features that help optimize your brushing 
routine. These include a built-in two-minute timer and a pressure sensor that 
alerts you if you're brushing too hard.

Overall, the AeroGlide UltraSlim Smart Toothbrush by Boie is a highly advanced 
and effective toothbrush that is perfect for anyone looking to enhance their 
oral hygiene. Its advanced sonic technology and smart features provide a deep 
and thorough clean, leaving your teeth fresh and healthy.
```

**问题分析：**
这个示例展示了典型的模型幻觉问题：
- 模型生成了一个听起来非常详细和专业的产品描述
- 描述包含了具体的技术细节（如"advanced sonic technology"、"high-frequency vibrations"）
- 描述看起来非常权威和可信
- **但实际上，模型可能并不知道这个产品的真实信息**，它只是基于产品名称和品牌，编造了一个听起来合理但可能不准确的描述

**关键教训：**
- 模型会自信地生成信息，即使它并不确定这些信息是否真实
- 对于特定产品、人物或事件的查询，模型可能会编造细节
- 在实际应用中，应该要求模型从提供的文档或可靠来源中查找信息，而不是依赖其训练数据

### Reducing hallucinations（减少幻觉）

**核心策略：** First find relevant information, then answer the question based on the relevant information.

**中文：** 先找到相关信息，然后基于相关信息回答问题。

**实施方法：**
1. **两步法**：
   - 第一步：让模型从提供的文档或数据源中查找相关信息
   - 第二步：基于找到的相关信息来回答问题

2. **使用分隔符**：
   - 将相关信息用分隔符（如三重引号、XML标签等）明确标识
   - 要求模型只基于分隔符内的信息回答问题

3. **引用来源**：
   - 要求模型在回答时引用信息来源
   - 如果找不到相关信息，要求模型明确说明

**优势：**
- 减少模型编造信息的可能性
- 提高答案的准确性和可靠性
- 让模型基于实际数据而非训练数据中的记忆来回答

---

## 迭代式提示词开发（Iterative Prompt Development）

### 迭代循环流程

提示词开发是一个迭代过程，通常遵循以下循环：

1. **Idea（想法）** → 2. **Implementation (code/data) Prompt（实现提示词）** → 3. **Experimental result（实验结果）** → 4. **Error Analysis（错误分析）** → 回到 **Idea**

**流程说明：**
- 从想法开始，编写提示词并实现
- 运行实验，观察结果
- 分析错误，找出问题所在
- 改进想法和提示词
- 重复上述过程，直到获得满意的结果

### 提示词开发指南（Prompt Guidelines）

1. **Be clear and specific（清晰且具体）**
   - 确保提示词表达清晰，指令具体明确
   - 使用之前学到的原则和策略

2. **Analyze why result does not give desired output（分析为什么结果没有给出期望的输出）**
   - 仔细检查模型输出与期望输出的差异
   - 识别问题所在：是指令不够清晰？还是缺少必要的上下文？
   - 理解模型为什么会产生这样的输出

3. **Refine the idea and the prompt（改进想法和提示词）**
   - 基于错误分析，改进初始想法
   - 优化提示词，使其更加清晰和具体
   - 可能需要添加更多细节、示例或约束条件

4. **Repeat（重复）**
   - 重复迭代过程，不断改进
   - 每次迭代都应该比上一次更接近目标

**关键要点：**
- 提示词开发不是一次性的工作，而是一个持续改进的过程
- 第一次尝试很少能获得完美结果，需要多次迭代优化
- 通过分析错误和不断改进，可以逐步提高提示词的效果

---

## 实际应用示例：将技术规格转换为营销文案

### 环境设置

**1. 加载 API Key：**
```python
import openai
import os
from dotenv import load_dotenv, find_dotenv

# 从 .env 文件加载环境变量
load_dotenv(find_dotenv())
openai.api_key = os.getenv("OPENAI_API_KEY")
```

**2. 定义辅助函数：**
```python
def get_completion(prompt, model="gpt-3.5-turbo"):
    """
    调用 OpenAI Chat Completion API
    
    参数:
    - prompt: 提示词字符串
    - model: 模型名称，默认为 "gpt-3.5-turbo"
    - temperature: 设置为 0，控制输出的随机性（0 = 更确定性，更少创造性）
    """
    messages = [{"role": "user", "content": prompt}]
    response = openai.ChatCompletion.create(
        model=model,
        messages=messages,
        temperature=0,  # 控制输出的随机性
    )
    return response.choices[0].message["content"]
```

### 示例：基于技术规格表生成产品描述

**任务：** 帮助营销团队基于技术规格表创建零售网站的产品描述。

**输入数据（技术规格表）：**
```python
fact_sheet_chair = """
OVERVIEW:
Part of a beautiful family of mid-century inspired office furniture, 
including filing cabinets, desks, bookcases, meeting tables, and more.
Several options of shell color and base finishes.
Available with plastic back and front upholstery (SWC-100) or full 
upholstery (SWC-110) in 10 fabric and 6 leather options.
Base finish options are: stainless steel, matte black, gloss white, or chrome.
Chair is available with or without armrests.
Suitable for home or business settings.
Qualified for contract use.

CONSTRUCTION:
5-wheel plastic coated aluminum base.
Pneumatic chair adjust for easy raise/lower action.

DIMENSIONS:
WIDTH: 53 CM | 20.87"
DEPTH: 51 CM | 20.08"
HEIGHT: 80 CM | 31.50"
SEAT HEIGHT: 44 CM | 17.32"
SEAT DEPTH: 41 CM | 16.14"

OPTIONS:
Soft or hard-floor caster options.
Two choices of seat foam densities: medium (1.8 lb/ft3) or high (2.8 lb/ft3).
Armless or 8 position PU armrests.

MATERIALS:
SHELL BASE GLIDER: Cast Aluminum with modified nylon PA6/PA66 coating.
Shell thickness: 10 mm.
SEAT: HD36 foam.

COUNTRY OF ORIGIN: Italy
"""
```

**提示词：**
```python
prompt = f"""
Your task is to help a marketing team create a description for a retail website 
of a product based on a technical fact sheet.

Write a product description based on the information provided in the technical 
specifications delimited by triple backticks.

Technical specifications: ```{fact_sheet_chair}```
"""

response = get_completion(prompt)
print(response)
```

**模型输出：**
```
Introducing our stunning mid-century inspired office chair, a perfect addition 
to any home or business setting. This chair is part of a beautiful family of 
office furniture, including filing cabinets, desks, bookcases, and meeting tables.

Customize your chair with several shell color and base finish options. Choose 
between plastic back and front upholstery (SWC-100) or full upholstery 
(SWC-110) in 10 fabric and 6 leather options. Base finish options include 
stainless steel, matte black, gloss white, or chrome. The chair is available 
with or without armrests and is qualified for contract use.

The chair features a 5-wheel plastic-coated aluminum base and a pneumatic adjust 
for easy raise/lower action. With a width of 53 cm, depth of 51 cm, height of 
80 cm, seat height of 44 cm, and seat depth of 41 cm, this chair provides 
ultimate comfort. Choose between soft or hard-floor caster options and two 
seat foam densities: medium (1.8 lb/ft3) or high (2.8 lb/ft3). Armrests are 
available as armless or an 8-position PU option.

Made with high-quality materials, the shell base glider is constructed from cast 
aluminum with modified nylon PA6/PA66 coating and has a shell thickness of 
10 mm. The seat is made of HD36 foam for comfort and durability.

Made in Italy, this chair is a perfect combination of style and functionality. 
Upgrade your workspace with this beautiful and practical office chair.
```

**应用要点：**
1. **使用分隔符**：使用三重反引号（```）将技术规格与指令分开
2. **明确任务**：清楚说明任务目标（为营销团队创建产品描述）
3. **结构化输入**：技术规格表以结构化格式提供，便于模型理解和提取信息
4. **输出转换**：模型成功将技术规格转换为吸引人的营销文案，保留了关键信息但使用了更营销化的语言

**这个示例展示了：**
- 如何将技术文档转换为面向消费者的内容
- 如何使用分隔符清晰标识输入数据
- 如何通过明确的指令指导模型完成特定任务

### 迭代改进示例：添加字数限制

**问题：** 第一次生成的描述可能太长，不适合某些应用场景（如产品卡片、社交媒体等）。

**改进版本的提示词：**
```python
prompt = f"""
Your task is to help a marketing team create a description for a retail website 
of a product based on a technical fact sheet.

Write a product description based on the information provided in the technical 
specifications delimited by triple backticks.

Use at most 50 words.

Technical specifications: ```{fact_sheet_chair}```
"""

response = get_completion(prompt)
print(response)
```

**模型输出（更简洁版本）：**
```
Introducing our mid-century inspired office chair, perfect for home or business 
settings. Available in a range of 11 colors and base finishes, with or without 
armrests. Choose from 10 fabric and 6 leather options for full or plastic 
upholstery. With a 5-wheel base and pneumatic chair adjust, it's both stylish 
and practical.
```

**改进点：**
- **添加约束条件**：通过"Use at most 50 words"限制输出长度
- **保持关键信息**：虽然更简洁，但仍保留了主要卖点（颜色选项、材质选择、功能特点）
- **适合不同场景**：短版本更适合产品卡片、社交媒体帖子等空间有限的场景

**迭代开发过程：**
1. **初始版本**：生成完整的产品描述（较长）
2. **分析需求**：发现需要更简洁的版本
3. **改进提示词**：添加字数限制约束
4. **验证结果**：检查输出是否符合要求

这展示了如何通过迭代改进提示词来满足不同的业务需求。

### 进一步迭代：添加更多具体要求和格式约束

**需求变化：** 现在需要为家具零售商创建技术性的产品描述，并且需要特定的输出格式。

**更复杂的提示词版本：**
```python
prompt = f"""
Your task is to help a marketing team create a
description for a retail website of a product based
on a technical fact sheet.

Write a product description based on the information
provided in the technical specifications delimited by
triple backticks.

The description is intended for furniture retailers,
so should be technical in nature and focus on the
materials the product is constructed from.

At the end of the description, include every 7-character
Product ID in the technical specification.

After the description, include a table that gives the
product's dimensions. The table should have two columns.
In the first column include the name of the dimension.
In the second column include the measurements in inches only.

Give the table the title 'Product Dimensions'.

Format everything as HTML that can be used in a website.
Place the description in a <div> element.

Technical specifications: ```{fact_sheet_chair}```
"""

response = get_completion(prompt)
print(response)
```

**这个版本的关键改进：**
1. **目标受众明确**：指定"for furniture retailers"（面向家具零售商）
2. **风格要求**：要求"technical in nature and focus on the materials"（技术性，关注材料）
3. **特定信息提取**：要求包含所有7字符的产品ID（如 SWC-100, SWC-110）
4. **结构化输出**：要求生成尺寸表格，包含两列（尺寸名称和英寸测量值）
5. **格式要求**：
   - 表格标题："Product Dimensions"
   - 输出格式：HTML
   - 描述放在 `<div>` 元素中

**迭代改进的层次：**
- **版本1**：基础产品描述（较长，通用风格）
- **版本2**：添加字数限制（简洁版本）
- **版本3**：添加目标受众、风格要求、特定信息提取、结构化输出和格式要求（技术性、HTML格式、包含表格）

**关键学习点：**
- 通过明确指定目标受众，可以改变输出的风格和重点
- 通过要求提取特定信息（产品ID），确保重要信息不被遗漏
- 通过指定输出格式（HTML、表格结构），可以直接用于实际应用
- 迭代改进是一个渐进的过程，可以根据需求逐步添加更多约束和要求

---

## 实际应用示例：情感分析和信息提取

### 示例1：基础情感分析

**任务：** 判断产品评论的情感倾向。

**输入数据（产品评论）：**
```python
lamp_review = """
Needed a nice lamp for my bedroom, and this one had \
additional storage and not too high of a price point. \
Got it fast.  The string to our lamp broke during the \
transit and the company happily sent over a new one. \
Came within a few days as well. It was easy to put \
together.  I had a missing part, so I contacted their \
support and they very quickly got me the missing piece! \
Lumina seems to me to be a great company that cares \
about their customers and products!!
"""
```

**提示词（版本1）：**
```python
prompt = f"""
What is the sentiment of the following product review, 
which is delimited with triple backticks?

Review text: '''{lamp_review}'''
"""

response = get_completion(prompt)
print(response)
```

**模型输出：**
```
The sentiment of the product review is positive.
```

**说明：** 基础版本成功识别了情感，但输出格式不够简洁。

### 示例2：改进版本 - 指定输出格式

**提示词（版本2）：**
```python
prompt = f"""
What is the sentiment of the following product review,
which is delimited with triple backticks?
Give your answer as a single word, either "positive" \
or "negative".

Review text: '''{lamp_review}'''
"""

response = get_completion(prompt)
print(response)
```

**模型输出：**
```
positive
```

**改进点：**
- **结构化输出**：要求输出单个词（"positive"或"negative"）
- **更简洁**：输出更简洁，便于程序处理
- **明确选项**：明确指定可能的输出值

### 示例3：提取情绪列表

**任务：** 识别评论者表达的情绪，最多5个，格式化为小写、逗号分隔的列表。

**提示词：**
```python
prompt = f"""
Identify a list of emotions that the writer of the \
following review is expressing. Include no more than \
five items in the list. Format your answer as a list of \
lower-case words separated by commas.

Review text: '''{lamp_review}'''
"""

response = get_completion(prompt)
print(response)
```

**模型输出：**
```
happy, satisfied, grateful, impressed, content
```

**应用要点：**
- **数量限制**：通过"no more than five items"控制输出长度
- **格式要求**：明确指定输出格式（小写、逗号分隔）
- **结构化输出**：便于后续处理和解析

### 示例4：判断特定情绪

**任务：** 判断评论是否表达愤怒。

**提示词：**
```python
prompt = f"""
Is the writer of the following review expressing anger?\
The review is delimited with triple backticks. \
Give your answer as either yes or no.

Review text: '''{lamp_review}'''
"""

response = get_completion(prompt)
print(response)
```

**模型输出：**
```
No
```

**应用场景：**
- 可以用于筛选负面评论
- 识别特定情绪状态
- 用于客户服务优先级排序

### 示例5：提取结构化信息（JSON格式）

**任务：** 从评论中提取商品名称和品牌，输出为JSON格式。

**提示词：**
```python
prompt = f"""
Identify the following items from the review text: 
- Item purchased by reviewer
- Company that made the item

The review is delimited with triple backticks.

Format your response as a JSON object with \
"Item" and "Brand" as the keys.
If the information isn't present, use "unknown" \
as the value.
Make your response as short as possible.

Review text: '''{lamp_review}'''
"""

response = get_completion(prompt)
print(response)
```

**模型输出：**
```json
{
  "Item": "lamp",
  "Brand": "Lumina"
}
```

**应用要点：**
1. **结构化输出**：要求JSON格式，便于程序处理
2. **明确键名**：指定JSON对象的键名（"Item"和"Brand"）
3. **处理缺失信息**：如果信息不存在，使用"unknown"作为值
4. **简洁性**：要求响应尽可能简短

**关键优势：**
- 相比传统的监督学习方法，使用提示词工程可以快速实现这些任务
- 无需训练专门的模型，只需编写合适的提示词
- 可以快速迭代和调整，适应不同的需求

**这些示例展示了：**
- 如何使用分隔符清晰标识输入文本
- 如何通过明确指令控制输出格式
- 如何提取结构化信息
- 如何根据不同需求调整提示词
- 提示词工程相比传统方法的效率优势

### 示例6：使用单个提示词提取多个字段

**任务：** 在一个提示词中同时提取多个信息字段，提高效率。

**提示词：**
```python
prompt = f"""
Identify the following items from the review text: 
- Sentiment (positive or negative)
- Is the reviewer expressing anger? (true or false)
- Item purchased by reviewer
- Company that made the item

The review is delimited with triple backticks.

Format your response as a JSON object with "Sentiment", "Anger", "Item" and "Brand" as the keys.
If the information isn't present, use "unknown" as the value.
Make your response as short as possible.
Format the Anger value as a boolean.

Review text: '''{lamp_review}'''
"""

response = get_completion(prompt)
print(response)
```

**模型输出：**
```json
{
  "Sentiment": "positive",
  "Anger": false,
  "Item": "lamp with additional storage",
  "Brand": "Lumina"
}
```

**关键优势：**
1. **一次提取多个字段**：使用单个提示词同时提取情感、愤怒状态、商品和品牌
2. **数据类型控制**：明确要求"Anger"值为布尔类型（true/false）
3. **结构化输出**：JSON格式便于程序处理
4. **效率提升**：相比分别调用多次API，这种方式更高效且成本更低

**应用场景：**
- 批量处理大量评论时，可以一次性提取所有需要的信息
- 减少API调用次数，降低成本和延迟
- 保持数据的一致性，因为所有信息来自同一次分析

**关键学习点：**
- 通过这种方式，可以使用单个提示从文本中提取多个字段
- 明确指定每个字段的数据类型（如布尔值）
- 合理组织提示词结构，列出所有需要提取的信息
- 统一处理缺失信息（使用"unknown"作为默认值）

**对比：**
- **分别提取**：需要4次API调用，成本高，延迟大
- **单次提取**：只需1次API调用，成本低，延迟小，数据一致性好

---

## 零样本学习（Zero-shot Learning）

### 概念说明

**零样本学习**：在机器学习中，这是指模型在没有针对特定任务进行训练或提供示例的情况下，仅根据指令就能执行任务的能力。

**在提示词工程中的应用：**
- 模型可以根据自然语言指令直接执行新任务
- 无需提供示例（few-shot）或进行微调（fine-tuning）
- 仅通过清晰的指令就能理解任务要求

### 示例：主题分类（Topic Classification）

**任务：** 判断文本中是否包含特定主题列表中的每个主题。

**提示词：**
```python
prompt = f"""
Determine whether each item in the following list of topics is a topic in the text below, 
which is delimited with triple backticks. Give your answer as list with 0 or 1 for each topic.

List of topics: {", ".join(topic_list)}
Text sample: '''{story}'''
"""

response = get_completion(prompt)
print(response)
```

**模型输出示例：**
```
nasa: 1
local government: 0
engineering: 0
employee satisfaction: 1
federal government: 1
```

**关键特点：**
1. **无需训练数据**：模型没有针对这个特定的主题分类任务进行训练
2. **无需示例**：提示词中没有提供任何示例（这是零样本，不是少样本）
3. **仅凭指令**：模型仅根据自然语言指令理解任务要求
4. **结构化输出**：输出格式明确（0或1），便于后续处理

**应用场景：**
- 文本分类：判断文本属于哪些类别
- 内容审核：检查文本是否包含特定主题
- 信息检索：快速筛选相关文档
- 标签生成：为文本自动添加标签

**优势：**
- **快速部署**：无需收集训练数据或进行模型训练
- **灵活性强**：可以轻松调整主题列表，无需重新训练
- **成本低**：相比训练专门模型，使用提示词工程成本更低
- **适应性强**：可以快速适应新的分类任务

**与少样本学习（Few-shot Learning）的区别：**
- **零样本（Zero-shot）**：不提供任何示例，仅凭指令
- **少样本（Few-shot）**：提供少量示例，让模型学习模式

**关键学习点：**
- 零样本学习展示了大型语言模型的强大泛化能力
- 通过清晰的指令，模型可以理解并执行新任务
- 这种方法特别适合快速原型开发和探索性任务
- 对于复杂任务，可能需要结合少样本学习或微调来获得更好效果

---

## 转换（Transforming）

转换任务是指将输入文本转换为不同的格式、风格或语言，这是提示词工程的另一个重要应用领域。

### 示例：语法检查和校对（Proofreading and Correction）

**任务：** 批量校对和更正包含语法错误、拼写错误和同音词误用的文本。

**输入数据（包含错误的句子列表）：**
```python
text = [
    "The girl with the black and white puppies have a ball.",  # 语法错误：have 应该是 has
    "Yolanda has her notebook.",  # ok
    "Its going to be a long day. Does the car need it's oil cha",  # 错误：Its 应该是 It's, it's 应该是 its, 句子不完整
    "Their goes my freedom. There going to bring they're suitca",  # 多个错误：Their/There/They're 混淆，句子不完整
    "Your going to need you're notebook.",  # 同音词错误：Your/You're 混淆
    "That medicine effects my ability to sleep. Have you heard",  # 错误：effects 应该是 affects
    "This phrase is to cherck chatGPT for speling abilitty"  # 拼写错误：cherck, speling, abilitty
]
```

**提示词：**
```python
for t in text:
    prompt = f"Proofread and correct: ```{t}```"
    response = get_completion(prompt)
    print(response)
```

**模型输出：**
```
The girl with the black and white puppies has a ball.
There are no errors in this sentence.

It's going to be a long day. Does the car need its oil changed?

There goes my freedom. They're going to bring their suitcases.

You're going to need your notebook.

That medicine affects my ability to sleep. Have you heard of the butterfly effect?

This phrase is to check ChatGPT for spelling ability.
```

**关键功能：**
1. **语法错误修正**：自动识别并修正语法错误（如主谓一致问题）
2. **拼写错误修正**：纠正拼写错误（如 "cherck" → "check", "speling" → "spelling"）
3. **同音词纠正**：区分并正确使用同音词（如 "Its/It's", "Your/You're", "Their/There/They're"）
4. **句子补全**：自动补全不完整的句子
5. **上下文理解**：根据上下文选择正确的词汇（如 "effects" vs "affects"）

**改进提示词的建议：**
为了改进提示词，可以更明确地说明任务：
```python
prompt = f"Proofread and correct the following text: ```{t}```"
```

**应用场景：**
- 文档校对和编辑
- 邮件和消息的语法检查
- 学术论文的初步校对
- 内容创作的质量控制
- 批量文本处理和质量提升

**优势：**
- 快速批量处理多个文本
- 识别多种类型的错误（语法、拼写、同音词）
- 提供上下文相关的修正建议
- 可以处理不完整的文本

### 示例：根据情感定制回复（Conditional Response Based on Sentiment）

**任务：** 创建一个客户服务AI助手，根据客户评论的情感生成不同的回复邮件。

**提示词：**
```python
prompt = f"""
You are a customer service AI assistant.
Your task is to send an email reply to a valued customer.
Given the customer email delimited by ```, \
Generate a reply to thank the customer for their review.
If the sentiment is positive or neutral, thank them for \
their review.
If the sentiment is negative, apologize and suggest that \
they can reach out to customer service.
Make sure to use specific details from the review.
Write in a concise and professional tone.
Sign the email as `AI customer agent`.

Customer email: ```{review}```
Review sentiment: {sentiment}
"""

response = get_completion(prompt)
print(response)
```

**关键要素：**

1. **角色设定**：
   - "You are a customer service AI assistant" - 明确AI的角色

2. **任务目标**：
   - "send an email reply to a valued customer" - 明确任务

3. **条件逻辑**：
   - **积极或中性情感**：感谢客户的评论
   - **消极情感**：道歉并建议联系客户服务

4. **具体要求**：
   - 使用评论中的具体细节
   - 简洁专业的语气
   - 署名"AI customer agent"

5. **输入格式**：
   - 使用分隔符（```）标识客户邮件
   - 明确提供评论情感信息

**应用场景：**
- 自动客户服务回复
- 评论管理和响应
- 社交媒体互动
- 客户反馈处理

**优势：**
- **个性化响应**：根据情感调整回复内容
- **自动化处理**：可以批量处理大量客户评论
- **一致性**：确保所有回复都遵循相同的标准和语气
- **效率提升**：减少人工处理时间

**模型输出示例（消极情感）：**

当客户评论的情感为消极时，模型生成的回复：

```
Dear valued customer,

Thank you for taking the time to leave a review about your
recent purchase of the 17 piece system. We are sorry to hear 
that you experienced a price increase and that the quality 
of the product did not meet your expectations. We apologize 
for any inconvenience this may have caused.

If you have any further concerns or questions, please do not 
hesitate to reach out to our customer service team. They will 
be more than happy to assist you in any way they can.

Thank you again for your feedback and for choosing our product. 
We hope to have the opportunity to serve you better in the future.

Best regards,
AI customer agent
```

**输出分析：**
- ✅ **识别消极情感**：正确识别并处理消极评论
- ✅ **使用具体细节**：引用了评论中的具体产品信息（"17 piece system"）
- ✅ **道歉和解决方案**：提供了道歉并建议联系客服
- ✅ **专业语气**：保持了简洁专业的语调
- ✅ **正确署名**：按照要求署名为"AI customer agent"

**关键学习点：**
- 通过条件逻辑（if-then）让模型根据输入生成不同的输出
- 明确指定角色和任务目标，帮助模型理解上下文
- 使用分隔符清晰标识输入内容
- 要求使用具体细节，确保回复的相关性和个性化
- 指定语气和格式，确保输出符合业务需求

**改进建议：**
- 可以添加更多情感类别（如"非常积极"、"轻微消极"等）
- 可以为不同情感类别指定不同的回复模板
- 可以要求模型提取关键信息并生成摘要
- 可以添加多语言支持

---

## Temperature 参数

### 概念说明

**Temperature（温度）**：控制语言模型输出随机性的参数，影响模型从概率分布中选择下一个词的方式。

**工作原理：**
- Temperature 值越高，输出越随机、越多样化
- Temperature 值越低，输出越确定、越可预测
- Temperature = 0 时，模型总是选择概率最高的选项（完全确定性）

### 示例：不同 Temperature 值的影响

**场景：** 模型需要完成句子 "my favorite food is"

**概率分布：**
- Pizza: 53%
- Sushi: 30%
- Tacos: 5%
- 其他选项: 12%

**不同 Temperature 值的输出：**

#### Temperature = 0（完全确定性）
```
输出1: my favorite food is pizza
输出2: my favorite food is pizza
输出3: my favorite food is pizza
```
- **特点**：总是选择概率最高的选项（Pizza, 53%）
- **适用任务**：**For tasks that require reliability, predictability**（需要可靠性和可预测性的任务）
- **应用场景**：代码生成、数据提取、事实问答、信息分类

#### Temperature = 0.3（低随机性）
```
输出1: my favorite food is pizza
输出2: my favorite food is sushi
输出3: my favorite food is pizza
```
- **特点**：主要选择高概率选项，偶尔选择中等概率选项（Sushi, 30%）
- **应用场景**：需要一定多样性但仍要保持相关性的任务（如内容创作、翻译）

#### Temperature = 0.7（中等随机性）
```
输出1: my favorite food is tacos
输出2: my favorite food is sushi
输出3: my favorite food is pizza
```
- **特点**：输出更加多样化，可能选择低概率选项（Tacos, 5%）
- **适用任务**：**For tasks that require variety**（需要多样性的任务）
- **应用场景**：创意写作、头脑风暴、内容创作、故事生成

### Temperature 值的选择指南

| Temperature 值 | 特点 | 任务类型 | 适用场景 |
|---------------|------|---------|---------|
| 0 | 完全确定性，总是选择最高概率 | **需要可靠性和可预测性** | 代码生成、数据提取、事实问答、信息分类 |
| 0.1 - 0.3 | 低随机性，主要选择高概率选项 | 需要一致性但允许少量变化 | 翻译、摘要、技术文档、客户服务回复 |
| 0.4 - 0.7 | 中等随机性，平衡确定性和多样性 | **需要多样性** | 内容创作、邮件撰写、对话生成、创意写作 |
| 0.8 - 1.0 | 高随机性，输出更加多样化 | 需要高创意和探索性 | 头脑风暴、故事生成、实验性创作 |
| > 1.0 | 极高随机性，可能产生不相关输出 | 实验性应用 | 不推荐常规使用 |

**核心原则：**
- **低 Temperature（0-0.3）**：适用于需要**可靠性和可预测性**的任务
- **高 Temperature（0.7-1.0）**：适用于需要**多样性**的任务

### 关键要点

1. **Temperature = 0**：
   - 输出完全可预测
   - 适合需要一致性的任务
   - 在之前的示例中，我们通常使用 `temperature=0` 来获得确定性输出

2. **低 Temperature（0.1-0.3）**：
   - 保持高相关性
   - 引入少量变化
   - 适合大多数实际应用

3. **高 Temperature（0.7-1.0）**：
   - 增加创意和多样性
   - 可能产生意外但有趣的输出
   - 需要更多验证和筛选

4. **实际应用建议**：
   - **信息提取任务**：使用 `temperature=0` 确保一致性
   - **内容创作任务**：使用 `temperature=0.7` 增加多样性
   - **对话系统**：使用 `temperature=0.3-0.5` 平衡相关性和自然度

**注意**：即使概率很低（如5%），在较高的temperature值下，模型仍可能选择该选项，这展示了temperature如何影响概率分布的采样方式。

---

## Role（角色）

### 概念说明

在对话API中，**Role（角色）**用于标识消息的发送者和类型，帮助模型理解对话的上下文和结构。

**如果你曾经使用过ChatGPT的Web界面**，你可能会注意到对话的历史记录。在API层面，这些对话是通过角色系统来管理的。

### 三种主要角色

#### 1. System（系统角色）

**作用**：设置AI助手的行为、个性和指令

**特点**：
- 通常只在对话开始时使用一次
- 定义了AI的整体行为模式
- 可以设置角色、语气、风格等

**示例：**
```python
messages = [
    {"role": "system", "content": "You are an assistant that helps users with coding questions."}
]
```

#### 2. User（用户角色）

**作用**：表示用户的输入或查询

**特点**：
- 代表用户发送的消息
- 可以是问题、指令、请求等
- 在对话中可以有多个user消息

**示例：**
```python
messages = [
    {"role": "user", "content": "tell me a joke"}
]
```

#### 3. Assistant（助手角色）

**作用**：表示AI助手的回复

**特点**：
- 代表AI生成的响应
- 通常由模型生成，但也可以手动添加（用于few-shot示例）
- 在对话历史中保存，帮助模型理解上下文

**示例：**
```python
messages = [
    {"role": "assistant", "content": "Why did the chicken cross the road? To get to the other side!"}
]
```

### 完整的对话示例

```python
messages = [
    {"role": "system", "content": "You are an assistant that helps users with coding questions."},
    {"role": "user", "content": "tell me a joke"},
    {"role": "assistant", "content": "Why did the chicken cross the road? To get to the other side!"},
    {"role": "user", "content": "Can you explain Python decorators?"},
    # ... 更多消息
]
```

### 角色系统的工作流程

1. **System消息**：设置AI的行为和角色
2. **User消息**：用户发送输入
3. **Assistant消息**：AI生成回复
4. **循环**：User和Assistant消息交替出现，形成对话历史

### 应用场景

1. **设置AI角色**：
   ```python
   {"role": "system", "content": "You are a helpful math tutor."}
   ```

2. **多轮对话**：
   - 通过保存完整的消息历史，实现上下文理解
   - 模型可以记住之前的对话内容

3. **Few-shot学习**：
   - 可以在messages中添加示例对话
   - 使用user和assistant角色展示期望的交互模式

4. **角色扮演**：
   ```python
   {"role": "system", "content": "You are a customer service representative for a tech company."}
   ```

### 关键要点

- **System角色**：定义AI的整体行为和个性
- **User角色**：用户的输入和查询
- **Assistant角色**：AI的回复，也用于保存对话历史
- **消息顺序**：消息的顺序很重要，模型按照顺序处理
- **上下文保持**：通过保存完整的消息历史，模型可以理解多轮对话的上下文

**实际应用：**
- 在之前的示例中，我们使用简单的`prompt`字符串
- 在实际的对话API中，应该使用`messages`数组和角色系统
- 这样可以更好地管理对话历史和上下文

### 重要提示：上下文管理

**这意味着您必须提供当前对话中所有相关的消息，以供模型使用。**

模型没有记忆功能，每次调用API时，都需要提供完整的对话历史。这包括：
- 系统消息（设置角色和行为）
- 所有之前的用户消息
- 所有之前的助手回复

### 示例：构建OrderBot聊天机器人

**目标：** 构建一个自动化的披萨订购聊天机器人，名为"OrderBot"。

#### 1. 系统提示词和菜单定义

```python
import panel as pn  # GUI
pn.extension()
panels = []  # collect display

context = [{'role': 'system', 'content': """
You are OrderBot, an automated service to collect orders for a pizza restaurant.
You first greet the customer, then collect the order,
and then ask if it's a pickup or delivery.
You wait to collect the entire order, then summarize it and
check for a final time if the customer wants to add anything else.
If it's a delivery, you ask for an address.
Finally you collect the payment.
Make sure to clarify all options, extras and sizes to uniquely identify the item from the menu.
You respond in a short, very conversational friendly style.

The menu includes:
- pepperoni pizza 12.95, 10.00, 7.00
- cheese pizza 10.95, 9.25, 6.50
- eggplant pizza 11.95, 9.75, 6.75
- fries 4.50, 3.50
- greek salad 7.25
- toppings: extra cheese 2.00, mushrooms 1.50, sausage 3.00, canadian bacon 3.50, AI sauce 1.50, peppers 1.00
- drinks: coke 3.00, 2.00, 1.00, sprite 3.00, 2.00, 1.00, bottled water 5.00
"""}]
```

**关键要素：**
- **角色定义**：明确OrderBot的身份和任务
- **交互流程**：定义完整的订购流程（问候→收集订单→确认配送方式→总结→付款）
- **菜单信息**：在系统提示词中包含完整的菜单和价格
- **响应风格**：指定简短、友好、对话式的风格

#### 2. 消息收集函数

```python
def collect_messages(_):
    prompt = inp.value_input
    inp.value = ''
    context.append({'role': 'user', 'content': f"{prompt}"})
    response = get_completion_from_messages(context)
    context.append({'role': 'assistant', 'content': f"{response}"})
    panels.append(pn.Row('User:', pn.pane.Markdown(prompt, width=600)))
    panels.append(pn.Row('Assistant:', pn.pane.Markdown(response, width=600)))
    return pn.Column(*panels)
```

**功能说明：**
1. **获取用户输入**：从输入框获取用户消息
2. **添加到上下文**：将用户消息添加到`context`列表
3. **获取AI回复**：调用`get_completion_from_messages`生成回复
4. **保存回复**：将助手回复添加到`context`列表
5. **显示对话**：使用Panel库显示用户和助手的消息

**关键点：**
- **自动收集**：自动收集用户提示和助手响应，构建完整的对话历史
- **上下文累积**：每次交互都会将新消息添加到`context`，保持完整的对话历史
- **可视化**：使用Panel库创建可视化的聊天界面

#### 3. 用户界面设置

```python
inp = pn.widgets.TextInput(value="Hi", placeholder='Enter text here...')
button_conversation = pn.widgets.Button(name="Chat!")
interactive_conversation = pn.bind(collect_messages, button_conversation)
dashboard = pn.Column(
    inp,
    pn.Row(button_conversation),
    pn.panel(interactive_conversation, loading_indicator=True)
)
dashboard
```

**界面组件：**
- **文本输入框**：用户输入消息
- **聊天按钮**：触发消息收集和显示
- **对话面板**：显示完整的对话历史

#### 4. 对话示例

**用户界面显示：**
```
User: HI I would like to order a pizza

Assistant: Hello! Welcome to our pizza restaurant. What can I get for you today?

User: How much are they

Assistant: Sure! The pepperoni pizza is $12.95 for a large, $10.00 for a medium, 
and $7.00 for a small. The cheese pizza is $10.95 for a large, $9.25 for a medium, 
and $6.50 for a small. The eggplant pizza is $11.95 for a large, $9.75 for a medium, 
and $6.75 for a small.
```

### 关键学习点

1. **上下文管理**：
   - 必须提供完整的对话历史
   - 每次API调用都需要包含所有之前的消息
   - 通过`context`列表累积所有消息

2. **角色系统应用**：
   - **System角色**：定义AI的行为和知识（菜单、流程）
   - **User角色**：用户的输入
   - **Assistant角色**：AI的回复，也用于保存对话历史

3. **自动化流程**：
   - 自动收集用户输入和AI回复
   - 自动添加到上下文
   - 实现真正的多轮对话

4. **实际应用**：
   - 可以构建真实的聊天机器人应用
   - 可以处理复杂的多步骤任务（如订单收集）
   - 可以集成到Web应用或移动应用中

### 构建聊天机器人的步骤总结

1. **定义系统提示词**：设置角色、任务、知识库（如菜单）
2. **初始化上下文**：创建包含系统消息的`context`列表
3. **创建消息收集函数**：处理用户输入，调用API，保存回复
4. **构建用户界面**：使用Panel或其他UI库创建交互界面
5. **测试和迭代**：不断测试和改进提示词和交互流程

**这个示例展示了如何将提示词工程应用到实际的产品中，构建一个完整的、可用的聊天机器人应用。**

---

## 课程总结（Summary）

### 核心原则（Principles）

1. **Write clear and specific instructions（编写清晰且具体的指令）**
   - 使用分隔符（delimiters）
   - 要求结构化输出
   - 检查条件是否满足
   - 使用少样本提示（few-shot prompting）

2. **Give the model time to "think"（给模型时间思考）**
   - 指定完成任务的步骤
   - 指导模型先自己解决问题，再得出结论

### 迭代式提示词开发（Iterative Prompt Development）

- 提示词开发是一个迭代过程
- 从想法开始，实现提示词，观察结果，分析错误，改进提示词
- 第一次尝试很少能获得完美结果，需要多次迭代优化
- 通过分析错误和不断改进，可以逐步提高提示词的效果

### 核心能力（Capabilities）

1. **Summarizing（总结）**
   - 文本摘要
   - 提取关键信息
   - 不同长度和风格的摘要

2. **Inferring（推理）**
   - 情感分析
   - 主题分类
   - 信息提取
   - 零样本学习

3. **Transforming（转换）**
   - 语言翻译
   - 语法检查和校对
   - 格式转换
   - 风格转换

4. **Expanding（扩展）**
   - 内容扩展
   - 根据情感定制回复
   - 个性化内容生成

### 构建聊天机器人（Building a Chatbot）

- 使用角色系统（Role）：System、User、Assistant
- 管理对话上下文：必须提供完整的对话历史
- 自动收集消息：构建完整的对话历史
- 创建交互界面：使用Panel等库构建用户界面

### 重要概念

- **Temperature参数**：控制输出的随机性
  - 低temperature（0-0.3）：需要可靠性和可预测性的任务
  - 高temperature（0.7-1.0）：需要多样性的任务

- **模型限制**：
  - Hallucination（幻觉）：模型可能生成不真实的信息
  - 减少幻觉：先找到相关信息，然后基于相关信息回答问题

- **零样本学习**：模型可以根据指令执行新任务，无需训练或示例

### 实际应用

- 客户服务自动化
- 内容创作和编辑
- 数据提取和分析
- 多语言翻译
- 文本分类和情感分析
- 聊天机器人和对话系统

### 关键要点

1. **清晰性**：提示词越清晰，输出越准确
2. **迭代性**：提示词开发需要多次迭代改进
3. **上下文**：提供足够的上下文信息
4. **结构化**：使用分隔符和结构化输出提高可处理性
5. **实践性**：通过实际项目应用所学知识

**提示词工程是一个强大的工具，通过掌握这些原则和技巧，可以有效地利用大型语言模型解决各种实际问题。**
