#version 330 core

layout (location = 0) in vec2 position;
layout (location = 1) in vec2 uv;
layout (location = 2) in vec4 color;

uniform vec2 screen;

out vec2 texCoord;
out vec4 tint;

void main() {
	vec2 ndc = vec2(position.x / screen.x * 2.0 - 1.0, 1.0 - position.y / screen.y * 2.0);
	gl_Position = vec4(ndc, 0.0, 1.0);
	texCoord = uv;
	tint = color;
}
