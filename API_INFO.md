# DeepSeek API 使用指南

## 基本信息

| 参数 | 值 |
| --- | --- |
| base_url | https://api.deepseek.com |
| api_key | 需在[平台申请](https://platform.deepseek.com/api_keys) |

> 注意：出于与OpenAI兼容考虑，您也可以将`base_url`设置为`https://api.deepseek.com/v1`，但此处的`v1`与模型版本无关。

## 模型信息

- **deepseek-chat**: 已全面升级为DeepSeek-V3，接口不变
- **deepseek-reasoner**: DeepSeek最新推出的推理模型DeepSeek-R1

## API调用示例

### Python示例

```python
# 安装OpenAI SDK: pip3 install openai

from openai import OpenAI

client = OpenAI(api_key="<DeepSeek API Key>", base_url="https://api.deepseek.com")

response = client.chat.completions.create(
    model="deepseek-chat",
    messages=[
        {"role": "system", "content": "You are a helpful assistant"},
        {"role": "user", "content": "Hello"},
    ],
    stream=False
)

print(response.choices[0].message.content)
```

### Curl示例

```bash
curl https://api.deepseek.com/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <DeepSeek API Key>" \
  -d '{
        "model": "deepseek-chat",
        "messages": [
          {"role": "system", "content": "You are a helpful assistant."},
          {"role": "user", "content": "Hello!"}
        ],
        "stream": false
      }'
```

### Node.js示例

```javascript
// 安装OpenAI SDK: npm install openai

import OpenAI from "openai";

const openai = new OpenAI({
        baseURL: 'https://api.deepseek.com',
        apiKey: '<DeepSeek API Key>'
});

async function main() {
  const completion = await openai.chat.completions.create({
    messages: [{ role: "system", content: "You are a helpful assistant." }],
    model: "deepseek-chat",
  });

  console.log(completion.choices[0].message.content);
}

main();
```

## 特殊功能

### 多轮对话

DeepSeek `/chat/completions` API是一个"无状态"API，服务端不记录用户请求的上下文。要实现多轮对话，需要在每次请求时将所有对话历史拼接好后传递给API。

#### 实现方式

```python
from openai import OpenAI
client = OpenAI(api_key="<DeepSeek API Key>", base_url="https://api.deepseek.com")

# 保存对话历史的列表
messages = []

# 第一轮对话
messages.append({"role": "user", "content": "What's the highest mountain in the world?"})
response = client.chat.completions.create(
    model="deepseek-chat",
    messages=messages
)

# 将模型回复添加到历史中
messages.append(response.choices[0].message)
print(f"Messages Round 1: {messages}")

# 第二轮对话
messages.append({"role": "user", "content": "What is the second?"})
response = client.chat.completions.create(
    model="deepseek-chat",
    messages=messages
)

# 再次将模型回复添加到历史中
messages.append(response.choices[0].message)
print(f"Messages Round 2: {messages}")
```

#### 注意事项

在实现多轮对话时，需要注意以下几点：

1. **正确的消息格式**：每条消息需要包含`role`和`content`两个字段
2. **完整的对话历史**：每次请求都需要传递完整的对话历史
3. **消息顺序**：消息需要按时间顺序排列，确保对话逻辑连贯
4. **消息数量控制**：如果历史太长，可能会超过模型的输入长度限制，需要适当控制或截断

例如，在第二轮对话时，传递给API的消息数组结构如下：

```json
[
    {"role": "user", "content": "What's the highest mountain in the world?"},
    {"role": "assistant", "content": "The highest mountain in the world is Mount Everest."},
    {"role": "user", "content": "What is the second?"}
]
```

### JSON Output功能

DeepSeek提供JSON Output功能，确保模型输出合法的JSON字符串，方便后续逻辑解析。

#### 使用注意事项

1. 设置`response_format`参数为`{'type': 'json_object'}`
2. 用户传入的system或user prompt中必须含有`json`字样，并给出希望模型输出的JSON格式样例
3. 需合理设置`max_tokens`参数，防止JSON字符串被中途截断
4. 使用此功能时，API有概率返回空content，可尝试修改prompt缓解此问题

#### 示例代码

```python
import json
from openai import OpenAI

client = OpenAI(
    api_key="<your api key>",
    base_url="https://api.deepseek.com",
)

system_prompt = """
The user will provide some exam text. Please parse the "question" and "answer" and output them in JSON format. 

EXAMPLE INPUT: 
Which is the highest mountain in the world? Mount Everest.

EXAMPLE JSON OUTPUT:
{
    "question": "Which is the highest mountain in the world?",
    "answer": "Mount Everest"
}
"""

user_prompt = "Which is the longest river in the world? The Nile River."

messages = [{"role": "system", "content": system_prompt},
            {"role": "user", "content": user_prompt}]

response = client.chat.completions.create(
    model="deepseek-chat",
    messages=messages,
    response_format={
        'type': 'json_object'
    }
)

print(json.loads(response.choices[0].message.content))
```

输出结果：
```json
{
    "question": "Which is the longest river in the world?",
    "answer": "The Nile River"
}
```

## 其他可用功能

根据文档，DeepSeek API还支持以下功能：
- 对话前缀续写（Beta）
- FIM补全（Beta）
- Function Calling
- 上下文硬盘缓存
- 提示库

更多详细信息可访问[DeepSeek API文档](https://api-docs.deepseek.com/zh-cn/)。 