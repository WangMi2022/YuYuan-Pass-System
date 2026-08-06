/*
 * Adapted from fluidglass-ui (Apache-2.0):
 * https://github.com/csuyincs-creator/fluidglass-ui
 */
    attribute vec2 a_position;
    varying vec2 v_uv;
    void main(){
      v_uv=a_position*.5+.5;
      gl_Position=vec4(a_position,0.0,1.0);
    }
  
