package raft

import (
	"errors"
	"sync"
	"time"
)

type Role uint8
const ( Follower Role = iota; Candidate; Leader )
type LogEntry struct { Term uint64 `json:"term"`; Index uint64 `json:"index"`; Command []byte `json:"command"` }
type RequestVote struct { Term, CandidateID, LastLogIndex, LastLogTerm uint64 }
type VoteResponse struct { Term uint64; Granted bool }
type AppendEntries struct { Term, LeaderID, PrevLogIndex, PrevLogTerm, LeaderCommit uint64; Entries []LogEntry }
type AppendResponse struct { Term uint64; Success bool; MatchIndex uint64 }
type StateMachine interface { Apply([]byte) error }

type Node struct { mu sync.Mutex; id uint64; role Role; currentTerm uint64; votedFor uint64; log []LogEntry; commitIndex uint64; lastApplied uint64; leaderID uint64; peers int; votes int; electionTimeout time.Duration; lastContact time.Time; machine StateMachine }

func NewNode(id uint64, peers int, machine StateMachine) *Node { if id==0{id=1};if peers<1{peers=1};return &Node{id:id,role:Follower,peers:peers,electionTimeout:150*time.Millisecond,lastContact:time.Now(),machine:machine} }
func (n *Node) State()(Role,uint64,uint64){n.mu.Lock();defer n.mu.Unlock();return n.role,n.currentTerm,n.leaderID}
func (n *Node) StartElection(now time.Time) RequestVote { n.mu.Lock();defer n.mu.Unlock();n.role=Candidate;n.currentTerm++;n.votedFor=n.id;n.votes=1;n.lastContact=now;if n.votes>n.peers/2{n.role=Leader;n.leaderID=n.id};var li,lt uint64;if len(n.log)>0{li=n.log[len(n.log)-1].Index;lt=n.log[len(n.log)-1].Term};return RequestVote{Term:n.currentTerm,CandidateID:n.id,LastLogIndex:li,LastLogTerm:lt} }
func (n *Node) HandleVote(req RequestVote) VoteResponse { n.mu.Lock();defer n.mu.Unlock();if req.Term<n.currentTerm{return VoteResponse{Term:n.currentTerm}};if req.Term>n.currentTerm{n.currentTerm=req.Term;n.role=Follower;n.votedFor=0;n.leaderID=0};lastIndex,lastTerm:=uint64(0),uint64(0);if len(n.log)>0{lastIndex=n.log[len(n.log)-1].Index;lastTerm=n.log[len(n.log)-1].Term};upToDate:=req.LastLogTerm>lastTerm||(req.LastLogTerm==lastTerm&&req.LastLogIndex>=lastIndex);granted:=(n.votedFor==0||n.votedFor==req.CandidateID)&&upToDate;if granted{n.votedFor=req.CandidateID;n.lastContact=time.Now()};return VoteResponse{Term:n.currentTerm,Granted:granted} }
func (n *Node) RecordVote(resp VoteResponse) bool { n.mu.Lock();defer n.mu.Unlock();if resp.Term>n.currentTerm{n.currentTerm=resp.Term;n.role=Follower;n.votedFor=0;n.leaderID=0;return false};if n.role!=Candidate||resp.Term!=n.currentTerm||!resp.Granted{return n.role==Leader};n.votes++;if n.votes>n.peers/2{n.role=Leader;n.leaderID=n.id;return true};return false }
func (n *Node) AppendLocal(command []byte)(LogEntry,error){if len(command)==0{return LogEntry{},errors.New("command required")};n.mu.Lock();defer n.mu.Unlock();if n.role!=Leader{return LogEntry{},errors.New("not leader")};index:=uint64(len(n.log))+1;e:=LogEntry{Term:n.currentTerm,Index:index,Command:append([]byte(nil),command...)};n.log=append(n.log,e);return e,nil}
func (n *Node) HandleAppend(req AppendEntries) AppendResponse { n.mu.Lock();defer n.mu.Unlock();if req.Term<n.currentTerm{return AppendResponse{Term:n.currentTerm}};if req.Term>n.currentTerm{n.currentTerm=req.Term;n.votedFor=0};n.role=Follower;n.leaderID=req.LeaderID;n.lastContact=time.Now();if req.PrevLogIndex>uint64(len(n.log)){return AppendResponse{Term:n.currentTerm}};if req.PrevLogIndex>0&&n.log[req.PrevLogIndex-1].Term!=req.PrevLogTerm{return AppendResponse{Term:n.currentTerm}};for _,e:=range req.Entries{if e.Index==0{continue};if e.Index<=uint64(len(n.log)){if n.log[e.Index-1].Term!=e.Term{n.log=n.log[:e.Index-1];n.log=append(n.log,e)}}else{n.log=append(n.log,e)}};if req.LeaderCommit>n.commitIndex{n.commitIndex=min(req.LeaderCommit,uint64(len(n.log)))};return AppendResponse{Term:n.currentTerm,Success:true,MatchIndex:uint64(len(n.log))} }
func (n *Node) ApplyCommitted() error { for { n.mu.Lock();if n.lastApplied>=n.commitIndex{n.mu.Unlock();return nil};e:=n.log[n.lastApplied];n.lastApplied++;machine:=n.machine;n.mu.Unlock();if machine!=nil{if err:=machine.Apply(append([]byte(nil),e.Command...));err!=nil{n.mu.Lock();n.lastApplied--;n.mu.Unlock();return err}} } }
func (n *Node) Snapshot()[]LogEntry{n.mu.Lock();defer n.mu.Unlock();out:=make([]LogEntry,len(n.log));for i,e:=range n.log{out[i]=LogEntry{Term:e.Term,Index:e.Index,Command:append([]byte(nil),e.Command...)}};return out}
func min(a,b uint64)uint64{if a<b{return a};return b}
