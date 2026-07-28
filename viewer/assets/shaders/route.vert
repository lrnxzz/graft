#version 330 core

layout (location = 0) in vec3 position;
layout (location = 1) in float along;

uniform mat4 viewProjection;

out float pathDistance;

void main() {
	gl_Position = viewProjection * vec4(position, 1.0);
	pathDistance = along;
}
