

run:
	clear && \
		cd client && go build -o ./mainClient client.go && mv ./mainClient ./../mainClient && cd ..
		cd server && go build -o ./mainServer server.go && mv ./mainServer ./../mainServer && cd ..
		cd bot && go build -o ./mainBot bot.go && mv ./mainBot ./../mainBot && cd ..
		clear


