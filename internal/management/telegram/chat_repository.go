package telegram

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type ChatRepository interface {
	GetAll() []*Chat
	AddChat(chat Chat) error
	RemoveChat(chatId int64) error
	IsChatExist(chatId int64) bool
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
func (f *FileChatRepository) AddChat(chat Chat) error {
	if f.IsChatExist(chat.Id) {
		return nil
	}
	c := append(f.chats, &chat)
	err := f.serializeChats(c)
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

func (f *FileChatRepository) IsChatExist(chatId int64) bool {
	return f.GetChat(chatId) != nil
}

func (f *FileChatRepository) RemoveChat(chatId int64) error {
	if !f.IsChatExist(chatId) {
		return nil
	}
	chats := make([]*Chat, 0)
	for _, chat := range f.chats {
		if chat.Id == chatId {
			continue
		}
		chats = append(chats, chat)
	}
	f.chats = chats
	err := f.serializeChats(chats)
	if err != nil {
		return err
	}
	f.chats = chats
	return nil
}

func (f *FileChatRepository) serializeChats(chats []*Chat) error {
	bytes, _ := json.MarshalIndent(chats, "", " ")
	_ = ioutil.WriteFile(f.filePath, bytes, 0644)
	return nil
}
