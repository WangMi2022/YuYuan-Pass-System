/*
 * Adapted from fluidglass-ui (Apache-2.0):
 * https://github.com/csuyincs-creator/fluidglass-ui
 */
    precision highp float;
    varying vec2 v_uv;
    uniform vec2 u_resolution;
    uniform vec2 u_mouse;
    uniform vec2 u_mouseVelocity;
    uniform float u_mouseMix;
    uniform float u_time;
    uniform float u_speed;
    uniform float u_intensity;
    uniform float u_pointer;
    uniform float u_seed;
    uniform float u_surfaceOpacity;
    uniform vec3 u_colorA;
    uniform vec3 u_colorB;
    uniform vec3 u_colorC;

    float hash(vec2 p){
      p=fract(p*vec2(123.34,456.21));
      p+=dot(p,p+45.32);
      return fract(p.x*p.y);
    }
    float noise(vec2 p){
      vec2 i=floor(p),f=fract(p);
      f=f*f*(3.0-2.0*f);
      return mix(mix(hash(i),hash(i+vec2(1.0,0.0)),f.x),mix(hash(i+vec2(0.0,1.0)),hash(i+vec2(1.0,1.0)),f.x),f.y);
    }
    float fbm(vec2 p){
      float value=0.0;
      float amp=.53;
      mat2 rot=mat2(.80,-.60,.60,.80);
      for(int i=0;i<5;i++){
        value+=amp*noise(p);
        p=rot*p*2.02+vec2(17.13,9.27);
        amp*=.49;
      }
      return value;
    }
    float softBlob(vec2 p,vec2 center,float radius,float softness){
      return 1.0-smoothstep(radius-softness,radius+softness,length(p-center));
    }
    void main(){
      vec2 uv=v_uv;
      float aspect=u_resolution.x/max(1.0,u_resolution.y);
      vec2 p=(uv-.5)*vec2(aspect,1.0);
      vec2 mouse=(u_mouse-.5)*vec2(aspect,1.0);
      vec2 delta=p-mouse;
      float dist=length(delta);
      float mouseField=exp(-dist*dist*7.2)*u_mouseMix*u_pointer;
      vec2 normal=delta/max(dist,.035);
      vec2 tangent=vec2(-normal.y,normal.x);
      p+=normal*mouseField*.115+tangent*mouseField*(u_mouseVelocity.x-u_mouseVelocity.y)*.045;

      float t=u_time*u_speed;
      vec2 seedVec=vec2(u_seed*1.713,u_seed*.937);
      float w1=fbm(p*1.22+seedVec+vec2(t*.075,-t*.052));
      float w2=fbm(p*1.54-seedVec*.37+vec2(-t*.057,t*.064)+w1*.82);
      vec2 q=p+(vec2(w1,w2)-.5)*(.58*u_intensity);
      float broad=fbm(q*1.12+vec2(t*.041,-t*.033));
      float detail=fbm(q*2.18+vec2(-t*.083,t*.057)+broad*.95);
      float ribbon=.5+.5*sin(q.x*3.15+q.y*.76+detail*5.0+t*.25+u_seed);
      float colorMix=smoothstep(.16,.88,broad*.61+ribbon*.39);
      vec3 fluid=mix(u_colorA,u_colorB,colorMix);
      float shadow=smoothstep(.43,.84,detail*.69+(.5+.5*sin(q.y*4.2-q.x*.8-t*.17))*.31);
      fluid=mix(fluid,u_colorC,shadow*.74);

      float plume1=softBlob(p,vec2(aspect*.23+.12*sin(t*.08+u_seed),.16*cos(t*.11+u_seed)),.52,.38);
      float plume2=softBlob(p,vec2(aspect*.39+.10*cos(t*.07-u_seed),-.24+.11*sin(t*.09)),.43,.34);
      float haze=clamp(plume1*.72+plume2*.58,0.0,1.0);
      float reveal=smoothstep(.055,.735,uv.x+(.5-broad)*.27+.070*sin(uv.y*4.0+t*.12));
      reveal*=mix(.70,1.0,haze);
      reveal=clamp(reveal*u_intensity,0.0,1.0);

      float spec=pow(clamp(1.0-abs(detail-.52)*2.0,0.0,1.0),5.0)*reveal;
      float caustic=pow(clamp(.52+.48*sin((q.x-q.y)*5.2+detail*7.0-t*.18),0.0,1.0),7.0)*reveal;
      vec3 cyanGlow=mix(fluid,vec3(.34,1.0,.90),spec*.18+caustic*.09);
      cyanGlow*=.78+.25*haze;
      cyanGlow=mix(cyanGlow,cyanGlow*.70,u_surfaceOpacity*.34);
      float filament=smoothstep(.48,.86,detail)*reveal;
      float density=clamp(reveal*(.36+.48*haze)+filament*.22+mouseField*.28,0.0,1.0);
      float alpha=clamp(.035*haze+density*(.24+.50*u_intensity)+spec*.08+u_surfaceOpacity*density*.14,0.0,.92);
      float edgeFade=smoothstep(1.08,.70,length((uv-.5)*vec2(1.0,.92)));
      alpha*=edgeFade;
      cyanGlow=pow(max(cyanGlow,0.0),vec3(.94));
      gl_FragColor=vec4(cyanGlow,alpha);
    }
  
