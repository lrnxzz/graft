#version 330 core

in float pathDistance;

uniform float walked;
uniform float time;

out vec4 fragColor;

void main() {
	float phase = fract(pathDistance * 1.5 - time * 1.1);
	if (phase > 0.62) {
		discard;
	}

	vec3 ahead = vec3(0.95, 0.23, 0.19);
	vec3 behind = vec3(0.22, 0.87, 0.35);
	vec3 color = pathDistance < walked ? behind : ahead;

	float glow = 0.85 + 0.15 * sin(time * 3.0);
	fragColor = vec4(color * glow, 1.0);
}
