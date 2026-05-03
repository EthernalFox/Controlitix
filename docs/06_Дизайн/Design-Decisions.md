# Design Decisions - Controlitix

��� ������������� ������-�������, ���������� �� ����� � ��������� ������-�������.

## DD-001 (2026-04-26): Liquid glass �� chrome editor-ui

�������: ��������� ������ liquid glass (backdrop-filter blur, �������������� ����, soft shadows) �� chrome-����� ������� editor-ui: Header, Panel, Aside, Footer, Modal, Popover, Dropdown.

��������: ���� `docs/06_������/Design-Brief-editor-ui.md` (�. 10) ��������� glassmorphism. �� ������� ��������� � ��������������� ����-���� `design_handoff_editor_ui/Controlitix Design System.html` (������ 16-86) ��������� glass ������ �� chrome. �������� ������, ������ ������ � ������ �������� ����� �������� ��������.

�����������:
- �������� ������ ����������� ��������: glass-��� �� ����� backdrop ������� contrast-ratio >= 4.5:1.
- Mobile/low-end: ��������� FPS canvas-����� ��� ���������� backdrop-filter; ��� ��������� ��������� glass ����� `prefers-reduced-transparency: reduce`.
- ������ �������� ��� `backdrop-filter` �������� fallback �� solid `--bg-surface`.
