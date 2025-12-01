# 📑 Documentation Index

Hướng dẫn toàn bộ tài liệu của Call Center AI System.

---

## 🚀 Getting Started

**Mới bắt đầu?** → Bắt đầu từ đây:

1. **[SUMMARY.md](SUMMARY.md)** 📋
   - Tổng quan project
   - Features overview
   - Tech stack
   - Achievements

2. **[QUICKSTART.md](QUICKSTART.md)** ⚡
   - Setup trong 10 phút
   - 2 options: Manual hoặc Docker
   - Quick troubleshooting
   - Test procedures

3. **[README.md](README.md)** 📖
   - Complete documentation
   - Detailed installation guide
   - API reference
   - Database schema

---

## 📚 Documentation by Topic

### Installation & Setup
- **Quick Setup**: [QUICKSTART.md](QUICKSTART.md)
- **Detailed Setup**: [README.md](README.md) → Installation section
- **Automated Setup**: `setup.sh` script
- **Docker Setup**: [docker-compose.yml](docker-compose.yml)

### Architecture & Code
- **Project Structure**: [PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md)
  - File organization
  - Code architecture
  - Data flow
  - Design patterns
  
### Deployment
- **Production Deployment**: [DEPLOYMENT.md](DEPLOYMENT.md)
  - VPS/Cloud deployment
  - Docker production setup
  - Security configuration
  - Monitoring & logging
  - Backup strategies
  - CI/CD pipeline

### API & Development
- **API Documentation**: [README.md](README.md) → API Documentation section
- **Database Models**: See [models.py](models.py)
- **API Schemas**: See [schemas.py](schemas.py)
- **CRUD Operations**: See [crud.py](crud.py)

---

## 🔍 Quick Links by Task

### "Tôi muốn..."

#### Setup hệ thống lần đầu
→ [QUICKSTART.md](QUICKSTART.md)

#### Deploy lên production
→ [DEPLOYMENT.md](DEPLOYMENT.md)

#### Hiểu cách hệ thống hoạt động
→ [PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md)

#### Xem API endpoints
→ [README.md](README.md) → API Documentation

#### Tùy chỉnh kịch bản
→ [README.md](README.md) → Scenarios section  
→ [ai_service.py](ai_service.py) → ScenarioManager class

#### Debug lỗi
→ [QUICKSTART.md](QUICKSTART.md) → Troubleshooting  
→ [DEPLOYMENT.md](DEPLOYMENT.md) → Troubleshooting Production

#### Test hệ thống
→ Run: `python test_system.py`

#### Backup database
→ [DEPLOYMENT.md](DEPLOYMENT.md) → Backup Strategy

---

## 📖 Documentation Files

### Core Documentation (Read First)

| File | Size | Purpose | Read Time |
|------|------|---------|-----------|
| [SUMMARY.md](SUMMARY.md) | 8K | Project overview | 5 min |
| [QUICKSTART.md](QUICKSTART.md) | 7K | Quick setup guide | 10 min |
| [README.md](README.md) | 11K | Complete documentation | 20 min |

### Technical Documentation

| File | Size | Purpose | Read Time |
|------|------|---------|-----------|
| [PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md) | 12K | Architecture details | 15 min |
| [DEPLOYMENT.md](DEPLOYMENT.md) | 13K | Production deployment | 20 min |

### Reference Files

| File | Purpose |
|------|---------|
| [.env.example](.env.example) | Environment variables template |
| [requirements.txt](requirements.txt) | Python dependencies |
| [docker-compose.yml](docker-compose.yml) | Docker setup |

---

## 💻 Source Code Files

### Main Application

| File | Lines | Purpose |
|------|-------|---------|
| [main.py](main.py) | ~400 | FastAPI application & routes |
| [database.py](database.py) | ~50 | Database connection |
| [models.py](models.py) | ~150 | SQLAlchemy models |
| [schemas.py](schemas.py) | ~120 | Pydantic schemas |
| [crud.py](crud.py) | ~300 | Database operations |

### Services

| File | Lines | Purpose |
|------|-------|---------|
| [ai_service.py](ai_service.py) | ~350 | AI integration (OpenAI, Anthropic) |
| [tts_service.py](tts_service.py) | ~300 | Text-to-Speech services |

### Utilities

| File | Lines | Purpose |
|------|-------|---------|
| [init_scenarios.py](init_scenarios.py) | ~50 | Initialize default scenarios |
| [test_system.py](test_system.py) | ~200 | System testing suite |
| [setup.sh](setup.sh) | ~250 | Automated setup script |

---

## 🎯 Learning Path

### Beginner
1. Read [SUMMARY.md](SUMMARY.md) - Hiểu tổng quan
2. Follow [QUICKSTART.md](QUICKSTART.md) - Setup hệ thống
3. Test call flow - Gọi và test thử
4. Read [README.md](README.md) → Features - Hiểu features

### Intermediate
1. Read [PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md) - Hiểu architecture
2. Review source code - [main.py](main.py), [ai_service.py](ai_service.py)
3. Customize scenarios - Edit trong database
4. Explore API - Test các endpoints

### Advanced
1. Read [DEPLOYMENT.md](DEPLOYMENT.md) - Production setup
2. Implement monitoring - Setup Prometheus/Grafana
3. Scale system - Multiple instances
4. Customize & extend - Add new features

---

## 🔧 Common Tasks Guide

### Setup & Configuration

```bash
# Quick setup
./setup.sh

# Manual setup
python -m venv venv
source venv/bin/activate
pip install -r requirements.txt
python init_scenarios.py

# Docker setup
docker-compose up -d
docker-compose exec api python init_scenarios.py
```

### Running the System

```bash
# Development
python main.py

# Production
uvicorn main:app --host 0.0.0.0 --port 8000 --workers 4

# Docker
docker-compose up -d
```

### Testing

```bash
# Test all components
python test_system.py

# Test with API
python test_system.py --with-api

# Test specific endpoint
curl http://localhost:8000/
```

### Database Operations

```bash
# Initialize scenarios
python init_scenarios.py

# MySQL backup
mysqldump -u callcenter -p call_center_db > backup.sql

# Restore
mysql -u callcenter -p call_center_db < backup.sql
```

---

## 📞 API Quick Reference

### Voice Endpoints
```
POST /voice/incoming         # Twilio incoming call
POST /voice/process-speech   # Process user speech
POST /voice/status          # Call status callback
```

### Call Management
```
GET    /api/calls           # List calls
GET    /api/calls/{id}      # Get call detail
DELETE /api/calls/{id}      # Delete call
```

### Scenario Management
```
GET    /api/scenarios       # List scenarios
POST   /api/scenarios       # Create scenario
GET    /api/scenarios/{id}  # Get scenario
PUT    /api/scenarios/{id}  # Update scenario
DELETE /api/scenarios/{id}  # Delete scenario
```

### Analytics
```
GET /api/analytics/statistics?days=30
```

Full API docs: http://localhost:8000/docs

---

## 🆘 Troubleshooting Index

### Common Issues

**Database connection failed**
→ [QUICKSTART.md](QUICKSTART.md) → Troubleshooting → Database

**AI service not available**
→ [QUICKSTART.md](QUICKSTART.md) → Troubleshooting → AI Service

**TTS not working**
→ [QUICKSTART.md](QUICKSTART.md) → Troubleshooting → TTS

**Twilio webhook issues**
→ [QUICKSTART.md](QUICKSTART.md) → Troubleshooting → Twilio

**Production issues**
→ [DEPLOYMENT.md](DEPLOYMENT.md) → Troubleshooting Production

### Debug Commands

```bash
# Check services
python test_system.py

# Check logs
tail -f /var/log/callcenter/app.log

# Check database
mysql -u callcenter -p call_center_db

# Check processes
ps aux | grep uvicorn
netstat -tulpn | grep 8000
```

---

## 🔗 External Resources

### APIs & Services
- [Twilio Voice Documentation](https://www.twilio.com/docs/voice)
- [OpenAI API Documentation](https://platform.openai.com/docs)
- [Anthropic Claude Documentation](https://docs.anthropic.com)
- [Google Cloud TTS](https://cloud.google.com/text-to-speech)
- [ElevenLabs API](https://elevenlabs.io/docs)

### Frameworks & Libraries
- [FastAPI Documentation](https://fastapi.tiangolo.com/)
- [SQLAlchemy Documentation](https://docs.sqlalchemy.org/)
- [Pydantic Documentation](https://docs.pydantic.dev/)
- [Docker Documentation](https://docs.docker.com/)

### Deployment
- [Nginx Documentation](https://nginx.org/en/docs/)
- [Let's Encrypt](https://letsencrypt.org/docs/)
- [Supervisor Documentation](http://supervisord.org/)

---

## 📊 Project Statistics

- **Total Files**: 17 Python + 6 Documentation
- **Lines of Code**: ~2,200+
- **Documentation**: ~50 pages / 5,000+ lines
- **Dependencies**: 17 packages
- **Supported Languages**: Vietnamese (primary), English
- **API Endpoints**: 15+
- **Database Tables**: 4
- **Built-in Scenarios**: 4

---

## 🎓 Additional Resources

### Video Tutorials (Nếu có)
- Setup Guide
- Configuration Tutorial
- API Usage Examples
- Deployment Walkthrough

### Code Examples
- See [main.py](main.py) for FastAPI examples
- See [ai_service.py](ai_service.py) for AI integration
- See [tts_service.py](tts_service.py) for TTS usage

### Sample Scenarios
- Customer Support: Hỗ trợ khách hàng
- Sales: Tư vấn bán hàng
- Appointment: Đặt lịch hẹn
- Survey: Khảo sát ý kiến

---

## 🤝 Contributing

Want to contribute?
1. Read [PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md)
2. Review coding standards
3. Test your changes
4. Submit pull request

---

## 📝 Version History

**v1.0.0** (December 2024)
- Initial release
- Complete call center system
- Full documentation
- Docker support
- Production ready

---

## 💡 Quick Tips

💡 **Tip 1**: Luôn backup database trước khi update  
💡 **Tip 2**: Test scenarios trước khi deploy production  
💡 **Tip 3**: Monitor logs thường xuyên  
💡 **Tip 4**: Sử dụng Docker cho consistency  
💡 **Tip 5**: Keep API keys secure trong .env  

---

## 🎯 Your Next Step

**Chọn path của bạn:**

- 🚀 **New User**: Start with [QUICKSTART.md](QUICKSTART.md)
- 🔧 **Developer**: Read [PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md)
- 🌐 **DevOps**: Go to [DEPLOYMENT.md](DEPLOYMENT.md)
- 📊 **Manager**: Review [SUMMARY.md](SUMMARY.md)

---

**Need help?** Check the documentation first, then create a GitHub issue.

**Happy Coding! 🎉**
