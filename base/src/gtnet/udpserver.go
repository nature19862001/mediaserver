package gtnet

import (
	"fmt"
	"net"
	"runtime"
)

type GTUDPPacket struct {
	Buff  []byte
	Raddr *net.UDPAddr
}

type GTUdpServer struct {
	OnStart    func()
	OnStop     func()
	OnError    func(int, string)
	OnRecv     func(*GTUDPPacket)
	OnPreSend  func(*GTUDPPacket)
	onPostSend func(*GTUDPPacket, int)
	//OnSend func([]byte, *net.UDPAddr)

	conn     *GTUdpConn
	sendChan chan *GTUDPPacket
}

func NewUdpServer(ip string, port int) *GTUdpServer {
	return &GTUdpServer{conn: NewUdpConn(ip, port), sendChan: make(chan *GTUDPPacket, 1024)}
}

// func NewUdpServer(addr *net.UDPAddr) *UdpServer {
// 	return &UdpServer{IP: addr.IP, Port: addr.Port, addr: addr}
// }

func (g *GTUdpServer) Start() error {
	var err error
	err = g.conn.StartListen()

	if err != nil {
		if g.OnError != nil {
			g.OnError(1, "StartListen error:"+err.Error())
		}
		return err
	}

	g.startUDPRecv()
	g.startUDPSend()

	if g.OnStart != nil {
		g.OnStart()
	}

	return nil
}

func (g *GTUdpServer) Stop() error {
	var err error

	err = g.conn.Close()

	if err != nil {
		if g.OnError != nil {
			g.OnError(1, "Stop error:"+err.Error())
		}
		return err
	}

	if g.OnStop != nil {
		g.OnStop()
	}

	return nil
}

func (g *GTUdpServer) Send(packet *GTUDPPacket) {
	g.sendChan <- packet
}

func (g *GTUdpServer) startUDPRecv() {
	go func() {
		buffer := make([]byte, 10240)

		for {
			num, raddr, err := g.conn.Recv(buffer)
			if err != nil {
				if g.OnError != nil {
					g.OnError(1, "Recv error:"+err.Error())
				}
				// if g.OnError != nil {
				// 	g.OnError(g.conn, 2, "ReadFromUDP err:"+err.Error())
				// }
				continue
			}

			newbuf := make([]byte, num)
			copy(newbuf, buffer[0:num])
			//newbuf = append(newbuf, buffer[0:num]...)
			g.OnRecv(&GTUDPPacket{newbuf, raddr})
		}
	}()
}

func (g *GTUdpServer) startUDPSend() {
	var numCPU = runtime.NumCPU()

	for i := 0; i < numCPU; i++ {
		go func() {
			for packet := range g.sendChan {
				if g.OnPreSend != nil {
					g.OnPreSend(packet)
				}

				num, err := g.conn.Send(packet.Buff, packet.Raddr)
				if err != nil {
					fmt.Println("err Send:" + err.Error())
					if g.OnError != nil {
						g.OnError(1, "Send error:"+err.Error())
					}
					return
				}

				if g.onPostSend != nil {
					g.onPostSend(packet, num)
				}
			}
		}()
	}
}
