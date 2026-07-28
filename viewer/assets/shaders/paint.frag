#version 330 core

in vec4 tint;

out vec4 fragColor;

void main() {
	fragColor = tint;
	if (fragColor.a < 0.01) {
		discard;
	}
}
