#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
智能版聊天测试工具 - AI自主评估关系变化
"""

import requests
import json
import time
import sys
import re
from dataclasses import dataclass
from typing import List, Dict, Optional

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
    MAGENTA = '\033[95m'
    END = '\033[0m'
    BOLD = '\033[1m'

@dataclass
class RelationshipState:
    """关系状态"""
    stage: int = 1
    familiarity: float = 0.0
    trust: float = 0.0
    intimacy: float = 0.0
    total_messages: int = 0

    def to_dict(self):
        return {
            "stage": self.stage,
            "familiarity": round(self.familiarity, 1),
            "trust": round(self.trust, 1),
            "intimacy": round(self.intimacy, 1),
            "total_messages": self.total_messages
        }

class RelationshipEvaluator:
    """AI驱动的关系评估器"""
    
    STAGE_NAMES = {1: "陌生期", 2: "熟悉期", 3: "亲近期", 4: "暧昧期"}
    
    # 阶段升级阈值（基于生物学曲线）
    STAGE_THRESHOLDS = {
        2: {"familiarity": 25, "trust": 15},
        3: {"familiarity": 55, "trust": 45, "intimacy": 25},
        4: {"familiarity": 75, "trust": 65, "intimacy": 50}
    }
    
    def __init__(self):
        self.state = RelationshipState()
        self.conversation_history = []  # 记录对话历史
        self.base_prompt = self._load_evaluator_prompt()
    
    def _load_evaluator_prompt(self) -> str:
        """加载评估器提示词"""
        try:
            with open('system_prompts/evaluator.txt', 'r', encoding='utf-8') as f:
                return f.read().strip()
        except FileNotFoundError:
            print(f"{Colors.YELLOW}警告: 找不到evaluator.txt，使用默认提示词{Colors.END}")
            return "你是人际关系专家，基于生物学和心理学原理评估对话。"
    
    def evaluate(self, user_msg: str, ai_msg: str) -> Dict:
        """使用AI评估对话"""
        
        # 保存到历史
        self.conversation_history.append({"user": user_msg, "ai": ai_msg})
        
        # 只保留最近5轮
        if len(self.conversation_history) > 5:
            self.conversation_history = self.conversation_history[-5:]
        
        # 构建历史对话文本
        history_text = self._format_history()
        
        # 构建完整评估prompt
        prompt = f"{self.base_prompt}

" + f"""━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
【当前关系状态】
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

阶段: {self.STAGE_NAMES[self.state.stage]} (Stage {self.state.stage})
熟悉度: {self.state.familiarity:.1f}/100
信任度: {self.state.trust:.1f}/100
亲密度: {self.state.intimacy:.1f}/100
对话轮数: {self.state.total_messages}

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
【对话历史】（最近{len(self.conversation_history)}轮）
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

{history_text}

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
【评估任务】
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

基于以上信息和生物学原理，评估最新一轮对话对关系的影响。

输出JSON格式（仅JSON，无其他内容）:
{{
  "familiarity_change": 数字（可正可负，可以是小数，根据真实影响判断），
  "trust_change": 数字（可正可负，可以是小数），
  "intimacy_change": 数字（可正可负，可以是小数），
  "is_key_moment": true/false,
  "reason": "简短分析（不超过30字）"
}}

重要提示：
- 客观评估，不被当前分数锚定
- 关键时刻可以产生大幅跃升（符合多巴胺机制）
- 考虑阶段特征，但以对话质量为准
- 负面互动应给予负分"""
        
        try:
            # 调用AI评估
            response = self._call_ai_evaluator(prompt)
            
            # 解析JSON
            result = self._parse_evaluation(response)
            
            return result
        except Exception as e:
            print(f"{Colors.RED}评估失败: {e}{Colors.END}")
            # 降级到简单规则
            return self._fallback_evaluation(user_msg, ai_msg)
    
    def _format_history(self) -> str:
        """格式化对话历史"""
        if not self.conversation_history:
            return "（暂无历史对话）"
        
        formatted = []
        for i, conv in enumerate(self.conversation_history, 1):
            formatted.append(f"第{i}轮:")
            formatted.append(f"  用户: {conv['user']}")
            formatted.append(f"  AI: {conv['ai']}")
        
        return "\n".join(formatted)
    
    def _call_ai_evaluator(self, prompt: str) -> str:
        """调用AI评估器"""
        url = f"{API_BASE_URL}/v1/chat/completions"
        headers = {
            "Authorization": f"Bearer {API_KEY}",
            "Content-Type": "application/json"
        }
        
        data = {
            "model": MODEL,
            "messages": [{"role": "user", "content": prompt}],
            "max_tokens": 200,
            "temperature": 0.3  # 评估用低温度，更稳定
        }
        
        response = requests.post(url, headers=headers, json=data, timeout=30)
        response.raise_for_status()
        result = response.json()
        return result['choices'][0]['message']['content']
    
    def _parse_evaluation(self, response: str) -> Dict:
        """解析AI返回的评估结果"""
        # 尝试提取JSON
        json_match = re.search(r'\{[^}]+\}', response, re.DOTALL)
        if json_match:
            json_str = json_match.group()
            result = json.loads(json_str)
            return result
        else:
            raise ValueError("无法解析AI返回的JSON")
    
    def _fallback_evaluation(self, user_msg: str, ai_msg: str) -> Dict:
        """降级评估（简单规则）"""
        user_len = len(user_msg)
        
        if user_len > 20:
            return {
                "familiarity_change": 5,
                "trust_change": 3,
                "intimacy_change": 1,
                "is_key_moment": False,
                "reason": "使用简单规则评估"
            }
        else:
            return {
                "familiarity_change": 2,
                "trust_change": 0,
                "intimacy_change": 0,
                "is_key_moment": False,
                "reason": "简短对话"
            }
    
    def update_state(self, evaluation: Dict):
        """更新关系状态"""
        self.state.familiarity += evaluation.get("familiarity_change", 0)
        self.state.trust += evaluation.get("trust_change", 0)
        self.state.intimacy += evaluation.get("intimacy_change", 0)
        
        # 限制在0-100范围
        self.state.familiarity = max(0, min(100, self.state.familiarity))
        self.state.trust = max(0, min(100, self.state.trust))
        self.state.intimacy = max(0, min(100, self.state.intimacy))
        
        self.state.total_messages += 1
        
        # 检查是否升级
        self._check_stage_upgrade()
    
    def _check_stage_upgrade(self):
        """检查阶段升级"""
        current_stage = self.state.stage
        
        for stage in range(2, 5):
            if stage > current_stage:
                threshold = self.STAGE_THRESHOLDS.get(stage, {})
                can_upgrade = all([
                    self.state.familiarity >= threshold.get("familiarity", 0),
                    self.state.trust >= threshold.get("trust", 0),
                    self.state.intimacy >= threshold.get("intimacy", 0)
                ])
                
                if can_upgrade:
                    self.state.stage = stage
                    print(f"\n{Colors.MAGENTA}{'='*60}{Colors.END}")
                    print(f"{Colors.MAGENTA}🎉 关系升级！ {self.STAGE_NAMES[stage]}{Colors.END}")
                    print(f"{Colors.MAGENTA}{'='*60}{Colors.END}\n")
                    break
    
    def get_stage_prompt(self) -> str:
        """获取当前阶段提示词"""
        stage_map = {1: "stranger", 2: "familiar", 3: "close", 4: "intimate"}
        stage_file = f"system_prompts/stage_{self.state.stage}_{stage_map[self.state.stage]}.txt"
        
        try:
            with open(stage_file, 'r', encoding='utf-8') as f:
                content = f.read().strip()
                # 注入当前分数
                return content.replace(
                    "系统分析：",
                    f"系统分析：当前分数 [熟悉{self.state.familiarity:.1f} 信任{self.state.trust:.1f} 亲密{self.state.intimacy:.1f}] - "
                )
        except FileNotFoundError:
            return f"<RELATIONSHIP_STATE>当前阶段: Stage {self.state.stage}</RELATIONSHIP_STATE>"
    
    def get_status_display(self) -> str:
        """状态显示"""
        return (
            f"{Colors.CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━{Colors.END}\n"
            f"{Colors.CYAN}关系: {self.STAGE_NAMES[self.state.stage]} (Stage {self.state.stage}) | "
            f"对话: {self.state.total_messages}轮{Colors.END}\n"
            f"{Colors.CYAN}熟悉度: {self.state.familiarity:.1f}/100 | "
            f"信任度: {self.state.trust:.1f}/100 | "
            f"亲密度: {self.state.intimacy:.1f}/100{Colors.END}\n"
            f"{Colors.CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━{Colors.END}"
        )

def load_base_prompt() -> str:
    """加载基础提示词"""
    try:
        with open('system_prompts/base.txt', 'r', encoding='utf-8') as f:
            return f.read().strip()
    except FileNotFoundError:
        print(f"{Colors.RED}错误: 找不到 system_prompts/base.txt{Colors.END}")
        sys.exit(1)

def call_ai(messages: List[Dict]) -> str:
    """调用主AI"""
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
    except Exception as e:
        return f"API错误: {e}"

def send_message_with_split(text: str):
    """分段发送"""
    parts = text.split('</>') if '</>' in text else [text]
    
    for i, part in enumerate(parts):
        part = part.strip()
        if not part:
            continue
        
        if i > 0:
            delay = 1 if len(part) < 10 else (2 if len(part) < 30 else 3)
            print(f"{Colors.CYAN}[延迟 {delay}s...]{Colors.END}")
            time.sleep(delay)
        
        print(f"{Colors.GREEN}AI: {part}{Colors.END}")

def main():
    """主函数"""
    print(f"{Colors.BOLD}{Colors.HEADER}{'='*60}{Colors.END}")
    print(f"{Colors.BOLD}{Colors.HEADER}  智能关系评估系统 - AI驱动{Colors.END}")
    print(f"{Colors.BOLD}{Colors.HEADER}{'='*60}{Colors.END}\n")
    
    base_prompt = load_base_prompt()
    evaluator = RelationshipEvaluator()
    
    print(f"{Colors.CYAN}✓ AI评估系统已启动{Colors.END}")
    print(f"{Colors.YELLOW}命令: 'quit'退出 | 'status'查看状态 | 'clear'重置{Colors.END}\n")
    print(evaluator.get_status_display())
    print()
    
    messages = []
    
    while True:
        try:
            user_input = input(f"{Colors.BLUE}你: {Colors.END}").strip()
            
            if not user_input:
                continue
            
            if user_input.lower() in ['quit', 'exit', 'q']:
                print(f"\n{Colors.YELLOW}再见！{Colors.END}")
                break
            
            if user_input.lower() == 'status':
                print("\n" + evaluator.get_status_display() + "\n")
                continue
            
            if user_input.lower() == 'clear':
                messages = []
                evaluator = RelationshipEvaluator()
                print(f"{Colors.CYAN}✓ 已重置{Colors.END}\n")
                print(evaluator.get_status_display())
                print()
                continue
            
            # 构建完整提示词
            stage_prompt = evaluator.get_stage_prompt()
            full_prompt = f"{base_prompt}\n\n{stage_prompt}"
            
            current_messages = [{"role": "system", "content": full_prompt}] + messages
            current_messages.append({"role": "user", "content": user_input})
            
            # 调用主AI
            print(f"{Colors.CYAN}[AI思考中...]{Colors.END}")
            ai_response = call_ai(current_messages)
            print('\033[A\033[K', end='')
            
            # 分段显示
            send_message_with_split(ai_response)
            
            # AI评估关系变化
            print(f"{Colors.YELLOW}[评估中...]{Colors.END}", end='', flush=True)
            evaluation = evaluator.evaluate(user_input, ai_response)
            print('\r\033[K', end='')  # 清除评估提示
            
            # 更新状态
            evaluator.update_state(evaluation)
            
            # 显示变化
            changes = []
            for key in ['familiarity', 'trust', 'intimacy']:
                val = evaluation.get(f"{key}_change", 0)
                if val != 0:
                    sign = '+' if val > 0 else ''
                changes.append(f"{key[:3]}{sign}{val:.1f}")
            
            if changes:
                change_str = ", ".join(changes)
                reason = evaluation.get('reason', '')
                key_mark = " 🔥" if evaluation.get('is_key_moment') else ""
                print(f"{Colors.YELLOW}[{change_str}]{key_mark} {reason}{Colors.END}")
            
            # 保存历史
            messages.append({"role": "user", "content": user_input})
            messages.append({"role": "assistant", "content": ai_response})
            
            print()
            
        except KeyboardInterrupt:
            print(f"\n\n{Colors.YELLOW}已中断{Colors.END}")
            break
        except Exception as e:
            print(f"{Colors.RED}错误: {e}{Colors.END}\n")

if __name__ == "__main__":
    main()
