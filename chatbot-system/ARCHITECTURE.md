# Kiến trúc hệ thống AI Chatbot

## 1. Hexagonal Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         EXTERNAL WORLD                                   │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐ │
│  │ Browser  │  │  Mobile  │  │   API    │  │  MySQL   │  │  Claude  │ │
│  │  Client  │  │   App    │  │  Client  │  │    DB    │  │   API    │ │
│  └─────┬────┘  └─────┬────┘  └─────┬────┘  └─────┬────┘  └─────┬────┘ │
└────────┼─────────────┼─────────────┼─────────────┼─────────────┼────────┘
         │             │             │             │             │
         │WebSocket    │WebSocket    │HTTP         │SQL          │HTTP
         ▼             ▼             ▼             ▼             ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    INFRASTRUCTURE LAYER (Adapters)                       │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │  Input Adapters (Primary/Driving)                                │  │
│  │  ┌─────────────────┐        ┌─────────────────┐                 │  │
│  │  │  HTTP Handler   │        │ WebSocket Hub   │                 │  │
│  │  │  - REST API     │        │  - Real-time    │                 │  │
│  │  │  - Get History  │        │  - Pub/Sub      │                 │  │
│  │  │  - Get Convs    │        │  - Connections  │                 │  │
│  │  └─────────────────┘        └─────────────────┘                 │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│                                    │                                     │
│                                    ▼                                     │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │  Output Adapters (Secondary/Driven)                              │  │
│  │  ┌─────────────┐  ┌─────────────┐  ┌──────────────────────────┐ │  │
│  │  │  Database   │  │  Database   │  │    AI Providers          │ │  │
│  │  │  Repos      │  │  Repos      │  │  ┌──────────────────┐    │ │  │
│  │  │  - User     │  │  - Message  │  │  │  Factory         │    │ │  │
│  │  │  - Conv     │  │             │  │  │  (Strategy)      │    │ │  │
│  │  └─────────────┘  └─────────────┘  │  └────────┬─────────┘    │ │  │
│  │                                     │           │              │ │  │
│  │                                     │  ┌────────▼────────┐     │ │  │
│  │                                     │  │ ClaudeProvider  │     │ │  │
│  │                                     │  │ (Strategy Impl) │     │ │  │
│  │                                     │  └─────────────────┘     │ │  │
│  │                                     │  ┌─────────────────┐     │ │  │
│  │                                     │  │DeepSeekProvider │     │ │  │
│  │                                     │  │ (Strategy Impl) │     │ │  │
│  │                                     │  └─────────────────┘     │ │  │
│  │                                     └──────────────────────────┘ │  │
│  └──────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                      APPLICATION LAYER (Use Cases)                       │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                        ChatUseCase                                │  │
│  │  ┌────────────────┐  ┌────────────────┐  ┌─────────────────┐    │  │
│  │  │ SendMessage()  │  │ GetHistory()   │  │ SwitchAIModel() │    │  │
│  │  │  - Validate    │  │  - Check perms │  │  - Validate     │    │  │
│  │  │  - Permission  │  │  - Retrieve    │  │  - Update conv  │    │  │
│  │  │  - Save msg    │  │  - Return      │  │                 │    │  │
│  │  │  - Call AI     │  │                │  │                 │    │  │
│  │  └────────────────┘  └────────────────┘  └─────────────────┘    │  │
│  └──────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        DOMAIN LAYER (Core Business)                      │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │  Entities                                                         │  │
│  │  ┌────────┐  ┌────────┐  ┌──────────────┐  ┌──────────────┐     │  │
│  │  │  User  │  │Message │  │ Conversation │  │   (others)   │     │  │
│  │  │        │  │        │  │              │  │              │     │  │
│  │  └────────┘  └────────┘  └──────────────┘  └──────────────┘     │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │  Interfaces (Ports)                                               │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐   │  │
│  │  │ UserRepo     │  │ MessageRepo  │  │  AIProvider          │   │  │
│  │  │ Interface    │  │ Interface    │  │  Interface           │   │  │
│  │  └──────────────┘  └──────────────┘  │  - SendMessage()     │   │  │
│  │  ┌──────────────┐                    │  - GetModelName()    │   │  │
│  │  │ ConvRepo     │                    │  - IsAvailable()     │   │  │
│  │  │ Interface    │                    │                      │   │  │
│  │  └──────────────┘                    └──────────────────────┘   │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │  Business Rules                                                   │  │
│  │  - User.CanChatWith(targetUser)                                  │  │
│  │  - Manager can chat with n users                                 │  │
│  │  - Advisor can chat with 1 user                                  │  │
│  │  - User can chat with their manager/advisor                      │  │
│  └──────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────┘
```

## 2. Strategy Pattern - AI Provider

```
┌─────────────────────────────────────────────────────────────┐
│                    AIProvider Interface                      │
│                    (Strategy Interface)                      │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  + SendMessage(ctx, conversation, msg) -> response   │  │
│  │  + GetModelName() -> string                          │  │
│  │  + IsAvailable() -> bool                             │  │
│  └───────────────────────────────────────────────────────┘  │
└────────────────────────────┬────────────────────────────────┘
                             │
            ┌────────────────┼────────────────┐
            │                │                │
            ▼                ▼                ▼
   ┌────────────────┐ ┌────────────────┐ ┌────────────────┐
   │ClaudeProvider  │ │DeepSeekProvider│ │ OpenAIProvider │
   │(Concrete       │ │(Concrete       │ │(Future         │
   │ Strategy)      │ │ Strategy)      │ │ Strategy)      │
   ├────────────────┤ ├────────────────┤ ├────────────────┤
   │- apiKey        │ │- apiKey        │ │- apiKey        │
   │- httpClient    │ │- httpClient    │ │- httpClient    │
   │- modelName     │ │- modelName     │ │- modelName     │
   ├────────────────┤ ├────────────────┤ ├────────────────┤
   │+ SendMessage() │ │+ SendMessage() │ │+ SendMessage() │
   │  -> calls      │ │  -> calls      │ │  -> calls      │
   │  Anthropic API │ │  DeepSeek API  │ │  OpenAI API    │
   └────────────────┘ └────────────────┘ └────────────────┘
            │                │                │
            └────────────────┼────────────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │ProviderFactory  │
                    │                 │
                    │+ GetProvider()  │
                    │+ ListProviders()│
                    └─────────────────┘
```

## 3. Database Schema Relationships

```
┌──────────────────────┐
│       users          │
├──────────────────────┤
│ id (PK)             │◄────────┐
│ username            │         │
│ email               │         │ manager_id (FK)
│ role                │         │
│ manager_id (FK) ────┼─────────┘
│ advisor_id (FK) ────┼───┐
│ created_at          │   │
│ updated_at          │   │ advisor_id (FK)
└──────────┬───────────┘   │
           │               │
           │               │
     1     │               │     1
           │               │
      ┌────┼───────────────┘
      │    │
      │    │ n (users)
      │    │
      │    ▼
      │ ┌──────────────────────┐
      │ │   conversations      │
      │ ├──────────────────────┤
      │ │ id (PK)             │
      └─┤ user1_id (FK) ───────┼─┐
        │ user2_id (FK)        │ │
        │ is_ai_chat           │ │
        │ active_ai_model      │ │
        │ created_at           │ │
        │ updated_at           │ │
        └──────────┬───────────┘ │
                   │             │
                   │             │
              1    │             │
                   │             │
                   │ n           │
                   │             │
                   ▼             │
        ┌──────────────────────┐ │
        │      messages        │ │
        ├──────────────────────┤ │
        │ id (PK)             │ │
        │ conversation_id (FK)─┘ │
        │ sender_id ─────────────┘
        │ content              │
        │ message_type         │
        │ ai_model             │
        │ created_at           │
        └──────────────────────┘
```

## 4. User Relationships & Permissions

```
┌─────────────────────────────────────────────────────────────┐
│                    User Hierarchy                            │
└─────────────────────────────────────────────────────────────┘

     Manager1 (role: manager)
         │
         ├─────── manages (1:n) ────┐
         │                           │
         ▼                           ▼
    User1 (role: user)          User2 (role: user)
         │
         │
    advised by (1:1)
         │
         ▼
    Advisor1 (role: advisor)


Permission Matrix:
┌──────────┬─────────────┬─────────────┬─────────────┐
│  User    │ Can Chat    │ Relationship│ Cardinality │
│  Role    │ With        │ Type        │             │
├──────────┼─────────────┼─────────────┼─────────────┤
│ Manager  │ Managed     │ Manages     │   1 : n     │
│          │ Users       │             │             │
├──────────┼─────────────┼─────────────┼─────────────┤
│ Advisor  │ Assigned    │ Advises     │   1 : 1     │
│          │ User        │             │             │
├──────────┼─────────────┼─────────────┼─────────────┤
│ User     │ Manager OR  │ Managed by  │   n : 1     │
│          │ Advisor     │ Advised by  │   1 : 1     │
├──────────┼─────────────┼─────────────┼─────────────┤
│ Any      │ AI Chatbot  │ AI Chat     │   1 : 1     │
└──────────┴─────────────┴─────────────┴─────────────┘
```

## 5. WebSocket Flow

```
    Client                          Server
      │                               │
      │──── Connect WS ──────────────►│
      │    ws://host/ws?user_id=1     │
      │                               │
      │◄──── Connected ───────────────│
      │    Register to Hub            │
      │                               │
      │──── Send Message ────────────►│
      │  {                            │
      │    type: "send_message",      │
      │    payload: {                 │
      │      content: "Hello AI",     │
      │      ai_model: "claude"       │
      │    }                          │
      │  }                            │
      │                               │
      │                         ┌─────▼─────┐
      │                         │ChatUseCase│
      │                         │           │
      │                         │1.Save msg │
      │                         │2.Call AI  │
      │                         │3.Save AI  │
      │                         │  response │
      │                         └─────┬─────┘
      │                               │
      │◄──── User Message ────────────┤
      │  {                            │
      │    type: "message",           │
      │    message_type: "user",      │
      │    content: "Hello AI"        │
      │  }                            │
      │                               │
      │◄──── AI Response ─────────────┤
      │  {                            │
      │    type: "message",           │
      │    message_type: "ai",        │
      │    ai_model: "claude",        │
      │    content: "Hi! How can..."  │
      │  }                            │
      │                               │
      ▼                               ▼
```

## 6. Request Flow Example

```
HTTP Request: GET /api/conversations/1/history?user_id=1

1. HTTP Handler
   └─► ChatHandler.GetConversationHistory()
       │
2.     ├─► Parse parameters (conversation_id, user_id)
       │
3.     └─► Call Use Case
           └─► ChatUseCase.GetConversationHistory(ctx, userID, convID, limit, offset)
               │
4.             ├─► Get Conversation from ConversationRepo
               │   └─► Check user has access
               │
5.             └─► Get Messages from MessageRepo
                   │
6.                 └─► Return messages
                     │
7. HTTP Handler ◄───┘
   │
8. └─► JSON Response
       {
         "messages": [...],
         "count": 10
       }
```
