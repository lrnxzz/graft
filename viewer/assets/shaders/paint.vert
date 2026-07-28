#version 330 core

layout (location = 0) in vec3 position;
layout (location = 1) in vec4 color;

uniform mat4 viewProjection;

out vec4 tint;

void main() {
	gl_Position = viewProjection * vec4(position, 1.0);
	tint = color;
}
