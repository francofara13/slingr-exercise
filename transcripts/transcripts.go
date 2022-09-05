package transcripts

import "strings"

const audio1 = "razu you must hurry you must stay the path"

const audio2 = "last eve i paused beside the blacksmith's door and heard the anvil ring the vesper chime then " +
	"looking in i saw upon the floor old hammers worn with beating years of time how many anvils " +
	"had you had said i to wear and better all these hammers so just one said he and then with " +
	"twinkling eye the anvil was the hamerton and so i thought the envious world for ages skeptic " +
	"blows have beat upon yet though the noise of falling blows was heard the anvil is alarmed the " +
	"hammers gone"

const audio3 = "the tyger by william blake read by a levity tiger tiger burning bright in the forests of the night " +
	"what immortal hand or eye could frame thy fearful symmetry in what distant deeps or skies " +
	"burned the fire of thine eyes on what wings dare he aspire what the hand dare seize the fire and " +
	"what shoulder and what art could twist the sinews of thy heart and when thy heart began to beat " +
	"what dread hand and what dread feet what the hammer what the chain in what furniss thy brain " +
	"what the anvil what dread grasp dare it deadly terrors clasp when the stars threw down their " +
	"spears and watered heaven with their tears did he smile his work to see did he who made the " +
	"lamb make thee tiger tiger burning bright in the forests of the night what immortal hand or eye " +
	"dare frame thy fearful symmetry in the poem this recording is in the public domain"

const audio4 = "algeria will rise again and dragunnitum is the key algeria will rise again and dragunnitum is the " +
	"key algeria will rise again and dragunnitum is the key algeria will rise again and dragunnitum is " +
	"the key algeria will rise again and dragunnitum is the key algeria will rise again and dragunnitum " +
	"is the key algeria will rise again and dragunnitum is the key"

const audio5 = "come on chief its trash can not trash cannot that was a bad motivational quote just go get them " +
	"come on chief its trash can not trash cannot that was a bad motivational quote just go get them " +
	"come on chief its trash can not trash cannot that was a bad motivational quote just go get them " +
	"come on chief its trash can not trash cannot that was a bad motivational quote just go get them " +
	"come on chief its trash can not trash cannot that was a bad motivational quote just go get them"

func GetAudioTranscript(audioName string) string {
	audioNameFormatted := strings.Split(audioName, ".wav")
	switch audioNameFormatted[0] {
	case "audio1":
		return audio1
	case "audio2":
		return audio2
	case "audio3":
		return audio3
	case "audio4":
		return audio4
	case "audio5":
		return audio5
	default:
		return ""
	}
}
