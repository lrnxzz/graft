#version 330 core

in vec2 texCoord;
in vec4 tint;

uniform sampler2D icons;

out vec4 fragColor;

void main() {
	vec4 texel = texCoord.x < 0.0 ? vec4(1.0) : texture(icons, texCoord);
	fragColor = texel * tint;
	if (fragColor.a < 0.01) {
		discard;
	}
}
