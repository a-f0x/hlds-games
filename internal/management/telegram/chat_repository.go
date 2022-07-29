package telegram

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type ChatRepository interface {
	GetAll() []*Chat
	SaveChat(chat *Chat) error
	RemoveChat(chatId int64) error
	GetChat(chatId int64) *Chat
}

const (
	chatFileName = "chats.json"
)

type FileChatRepository struct {
	filePath string
	chats    []*Chat
}

func NewFileChatRepository(dataDirPath string) (*FileChatRepository, error) {
	f := &FileChatRepository{
		filePath: fmt.Sprintf("%s/%s", dataDirPath, chatFileName),
		chats:    make([]*Chat, 0),
	}
	chatsFile, err := os.OpenFile(f.filePath, os.O_CREATE|os.O_RDONLY, 0666)
	if err != nil {
		return nil, fmt.Errorf("fail to open or create file %s. %s", f.filePath, err.Error())
	}
	defer chatsFile.Close()

	stat, err := chatsFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("fail to read stats of file %s. %s", f.filePath, err.Error())
	}

	b := make([]byte, stat.Size())
	_, err = chatsFile.Read(b)
	if err != nil {
		return nil, fmt.Errorf("fail to read bytes from file %s. %s", f.filePath, err.Error())
	}
	if len(b) == 0 {
		return f, nil
	}
	err = json.Unmarshal(b, &f.chats)
	if err != nil {
		return nil, fmt.Errorf("fail to read chats json from file %s. %s", f.filePath, err.Error())
	}
	return f, nil
}

func (f *FileChatRepository) GetAll() []*Chat {
	c := make([]*Chat, len(f.chats))
	copy(c, f.chats)
	return c
}
func (f *FileChatRepository) SaveChat(chat *Chat) error {
	i := f.getChatIndex(chat.Id)
	if i != -1 {
		chatsCopy := make([]*Chat, len(f.chats))
		copy(chatsCopy, f.chats)
		chatsCopy[i] = chat
		err := f.flush(chatsCopy)
		if err != nil {
			return err
		}
		f.chats = chatsCopy
		return nil
	}
	c := append(f.chats, chat)
	err := f.flush(c)
	if err != nil {
		return err
	}
	f.chats = c
	return nil
}

func (f *FileChatRepository) GetChat(chatId int64) *Chat {
	for _, chat := range f.chats {
		if chat.Id == chatId {
			return chat
		}
	}
	return nil
}

func (f *FileChatRepository) getChatIndex(chatId int64) int {
	for i, c := range f.chats {
		if c.Id == chatId {
			return i
		}
	}
	return -1
}

func (f *FileChatRepository) RemoveChat(chatId int64) error {
	chats := make([]*Chat, 0)
	for _, chat := range f.chats {
		if chat.Id == chatId {
			continue
		}
		chats = append(chats, chat)
	}
	err := f.flush(chats)
	if err != nil {
		return err
	}
	f.chats = chats
	return nil
}

func (f *FileChatRepository) flush(chats []*Chat) error {
	bytes, _ := json.MarshalIndent(chats, "", " ")
	_ = ioutil.WriteFile(f.filePath, bytes, 0644)
	return nil
}
