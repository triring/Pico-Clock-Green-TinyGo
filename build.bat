cd Clock
# tinygo flash -target=pico -size=short .
tinygo build -o Clock.uf2 --target pico --size short .
dir *.uf2
cd ..

cd Pomodoro
# tinygo flash -target=pico -size=short .
tinygo build -o Pomodoro.uf2 --target pico --size short .
dir *.uf2
cd ..

cd StopWatch
# tinygo flash -target=pico -size=short .
tinygo build -o StopWatch.uf2 --target pico --size short .
dir *.uf2
cd ..

cd Telop
# tinygo flash -target=pico -size=short .
tinygo build -o Telop.uf2 --target pico --size short .
dir *.uf2
cd ..

cd Tetris
# tinygo flash -target=pico -size=short .
tinygo build -o Tetris.uf2 --target pico --size short .
dir *.uf2
cd ..

cd Timer
# tinygo flash -target=pico -size=short .
tinygo build -o Timer.uf2 --target pico --size short .
dir *.uf2
cd ..

