# 도토리 구동기 포크

기준: df-mc/dragonfly 1793369709ce. 원본 모듈 경로와 MIT 라이선스를 유지합니다.

- 도토리 개발 브랜치: `dotori`
- 초기 구동기 태그: `v0.11.6-dotori.1`
- 프로토콜 포크: [RICHMCBE/gophertunnel](https://github.com/RICHMCBE/gophertunnel)
- 기능별 비공개 모듈과 Windows/Linux 통합 빌드: [RICHMCBE/dotori-server](https://github.com/RICHMCBE/dotori-server) (조직 권한 필요)

이 저장소에는 구동기 호환 확장만 포함합니다. 게임 기능 소스, 운영 설정, 월드, 플레이어 데이터, 리소스팩 및 인증키는 포함하지 않습니다.

주요 변경은 PMMP 데이터/NBT 보존, 가상 상자와 카메라, 커스텀 아이템 및 지도, 비행 권한, 경험치/체력 표시, 월드 트랜잭션 훅, 스킨 메타데이터 보존, 채팅 발신자 처리입니다.
포크의 Actions는 구동기 테스트와 컴파일만 수행합니다. 비공개 기능이 포함된 실행 파일은 비공개 조립 저장소에서만 생성합니다.

업스트림 갱신은 `dotori` 브랜치에서 검토하고 회귀 테스트를 거친 뒤 새로운 `v*-dotori.*` 태그로 배포합니다. 기존 태그는 이동하지 않습니다.
