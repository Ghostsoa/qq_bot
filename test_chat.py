#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
简化版聊天测试工具
用于测试系统提示词效果
"""

import requests
import json
import time
import sys

# 配置
API_BASE_URL = "https://api.deepseek.com"
API_KEY = "sk-593692de98614e81baf15878043c30c9"
MODEL = "deepseek-chat"
MAX_TOKENS = 500
TEMPERATURE = 0.95

# 颜色输出
class Colors:
    HEADER = '\033[95m'
    BLUE = '\033[94m'
    CYAN = '\033[96m'
    GREEN = '\033[92m'
    YELLOW = '\033[93m'
    RED = '\033[91m'
    END = '\033[0m'
    BOLD = '\033[1m'

def load_system_prompt():
    """加载系统提示词"""
    try:
        with open('system_prompt.txt', 'r', encoding='utf-8') as f:
            return f.read().strip()
    except FileNotFoundError:
        print(f"{Colors.RED}错误: 找不到 system_prompt.txt 文件{Colors.END}")
        sys.exit(1)

def call_ai(messages):
    """调用 AI API"""
    url = f"{API_BASE_URL}/v1/chat/completions"
    headers = {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": "application/json"
    }
    
    data = {
        "model": MODEL,
        "messages": messages,
        "max_tokens": MAX_TOKENS,
        "temperature": TEMPERATURE
    }
    
    try:
        response = requests.post(url, headers=headers, json=data, timeout=30)
        response.raise_for_status()
        result = response.json()
        return result['choices'][0]['message']['content']
    except requests.exceptions.RequestException as e:
        return f"API 请求失败: {e}"
    except (KeyError, IndexError) as e:
        return f"解析响应失败: {e}"

def send_message_with_split(text):
    """模拟分段发送（带延迟）"""
    parts = text.split('</>') if '</>' in text else [text]
    
    for i, part in enumerate(parts):
        part = part.strip()
        if not part:
            continue
        
        # 第一条消息立即发送，后续消息延迟
        if i > 0:
            length = len(part)
            if length < 10:
                delay = 1
            elif length < 30:
                delay = 2
            else:
                delay = 3
            
            print(f"{Colors.CYAN}[延迟 {delay}s...]{Colors.END}")
            time.sleep(delay)
        
        print(f"{Colors.GREEN}AI: {part}{Colors.END}")

def main():
    """主函数"""
    print(f"{Colors.BOLD}{Colors.HEADER}{'='*60}{Colors.END}")
    print(f"{Colors.BOLD}{Colors.HEADER}  QQ Bot 提示词测试工具{Colors.END}")
    print(f"{Colors.BOLD}{Colors.HEADER}{'='*60}{Colors.END}\n")
    
    # 加载系统提示词
    system_prompt = load_system_prompt()
    print(f"{Colors.CYAN}✓ 系统提示词已加载{Colors.END}")
    print(f"{Colors.YELLOW}提示: 输入 'quit' 或 'exit' 退出，输入 'clear' 清空历史{Colors.END}\n")
    
    # 初始化对话历史（临时内存）
    messages = [{"role": "system", "content": system_prompt}]
    
    while True:
        try:
            # 获取用户输入
            user_input = input(f"{Colors.BLUE}你: {Colors.END}").strip()
            
            if not user_input:
                continue
            
            # 退出命令
            if user_input.lower() in ['quit', 'exit', 'q']:
                print(f"\n{Colors.YELLOW}再见！{Colors.END}")
                break
            
            # 清空历史命令
            if user_input.lower() == 'clear':
                messages = [{"role": "system", "content": system_prompt}]
                print(f"{Colors.CYAN}✓ 对话历史已清空{Colors.END}\n")
                continue
            
            # 添加用户消息到历史
            messages.append({"role": "user", "content": user_input})
            
            # 调用 AI
            print(f"{Colors.CYAN}[思考中...]{Colors.END}")
            ai_response = call_ai(messages)
            
            # 清除"思考中"提示
            print('\033[A\033[K', end='')
            
            # 保存 AI 回复到历史（完整内容）
            messages.append({"role": "assistant", "content": ai_response})
            
            # 分段显示 AI 回复
            send_message_with_split(ai_response)
            print()  # 空行分隔
            
        except KeyboardInterrupt:
            print(f"\n\n{Colors.YELLOW}程序已中断{Colors.END}")
            break
        except Exception as e:
            print(f"{Colors.RED}错误: {e}{Colors.END}\n")

if __name__ == "__main__":
    main()
