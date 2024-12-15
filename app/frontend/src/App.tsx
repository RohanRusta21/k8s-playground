import React, { useState, useEffect } from 'react';
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";


function App() {
  // Define the Todo interface
  interface Todo {
    uuid: string;
    title: string;
    description: string;
    completed: boolean;
  }

  const [todos, setTodos] = useState<Todo[]>([]);
  const [newTodo, setNewTodo] = useState({ title: '', description: '' });
  const [files, setFiles] = useState<string[]>([]);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);

  useEffect(() => {
    fetchTodos();
    fetchFiles();
  }, []);

  // Get base URL from environment variable
  const BASE_URL = process.env.REACT_APP_API_BASE_URL || 'http://localhost:8080/api';
  console.log('REACT_APP_API_BASE_URL:', process.env.REACT_APP_API_BASE_URL);
  console.log('BASE_URL:', BASE_URL);


  const fetchTodos = async () => {
    try {
      const response = await fetch(BASE_URL+'/todos');
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      const data: Todo[] = await response.json();
      setTodos(data);
    } catch (error) {
      console.error('Error fetching todos:', error);
    }
  };

  const fetchFiles = async () => {
    try {
      const response = await fetch(BASE_URL+'/files/list');
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      const data: string[] = await response.json();
      if (data !== null) {
        setFiles(data);
      }
      else {
        setFiles([]);
      }
    } catch (error) {
      console.error('Error fetching files:', error);
    }
  };

  const handleCreateTodo = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    try {
      const response = await fetch(BASE_URL+'/todos', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(newTodo)
      });
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      await fetchTodos();
      setNewTodo({ title: '', description: '' });
    } catch (error) {
      console.error('Error creating todo:', error);
    }
  };

  const handleEditTodo = async (uuid: string, updatedTodo: Partial<Todo>) => {
    try {
      setTodos((prevTodos) =>
        prevTodos.map((todo) =>
          todo.uuid === uuid ? { ...todo, ...updatedTodo } : todo
        )
      );
  
      const response = await fetch(BASE_URL+`/todos/${uuid}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(updatedTodo),
      });
  
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
  
      // Optionally refresh todos from server
      await fetchTodos();
    } catch (error) {
      console.error('Error updating todo:', error);
    }
  };
  

  const handleDeleteTodo = async (uuid: string) => {
    try {
      const response = await fetch(`${BASE_URL}/todos/${uuid}`, {
        method: 'DELETE'
      });
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      await fetchTodos();
    } catch (error) {
      console.error('Error deleting todo:', error);
    }
  };

  const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    if (!e.target.files || e.target.files.length === 0) {
      console.error('No file selected');
      return;
    }

    const file = e.target.files[0];
    setSelectedFile(file);
    const formData = new FormData();
    formData.append('file', file);

    try {
      const response = await fetch(BASE_URL+'/files/upload', {
        method: 'POST',
        body: formData
      });
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      await fetchFiles();
    } catch (error) {
      console.error('Error uploading file:', error);
    }
  };

  const handleFileDownload = async (fileName: string) => {
    try {
      const response = await fetch(`${BASE_URL}/files/download/${fileName}`);
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = fileName;
      a.click();
      window.URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Error downloading file:', error);
    }
  };

  const handleFileDelete = async (fileName: string) => {
    try {
        const response = await fetch(`${BASE_URL}/files/${fileName}`, {
            method: 'DELETE'
        });
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        // Immediately update the files list
        await fetchFiles();
    } catch (error) {
        console.error('Error deleting file:', error);
    }
};


  return (
    <div className="container mx-auto p-4">
      <Card className="mb-4">
        <CardHeader>
          <CardTitle>Create Todo</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleCreateTodo} className="space-y-4">
            <Input 
              placeholder="Todo Title" 
              value={newTodo.title}
              onChange={(e) => setNewTodo({ ...newTodo, title: e.target.value })}
            />
            <Input 
              placeholder="Description" 
              value={newTodo.description}
              onChange={(e) => setNewTodo({ ...newTodo, description: e.target.value })}
            />
            <Button type="submit">Create Todo</Button>
          </form>
        </CardContent>
      </Card>

      <Card className="mb-4">
        <CardHeader>
          <CardTitle>Todos</CardTitle>
        </CardHeader>
        <CardContent>
          {todos.map((todo) => (
            <div key={todo.uuid} className="flex items-center space-x-2 mb-2">
              <Checkbox 
                checked={todo.completed} 
                onCheckedChange={() => handleEditTodo(todo.uuid, { completed: !todo.completed })}
              />
              <span className="flex-grow">{todo.title} - {todo.description}</span>
              <Button onClick={() => handleDeleteTodo(todo.uuid)}>Delete</Button>
            </div>
          ))}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>File Management</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="mb-4">
            <input 
              type="file" 
              onChange={handleFileUpload} 
              className="mb-2"
            />
            {selectedFile && (
              <p className="text-sm text-gray-500">
                Selected file: {selectedFile.name}
              </p>
            )}
          </div>
          
          <div>
            <h3 className="text-lg font-semibold mb-2">Uploaded Files</h3>
            {files.length === 0 ? (
              <p className="text-gray-500">No files uploaded yet</p>
            ) : (
              <ul className="space-y-2">
    {files.map((fileName) => (
        <li 
            key={fileName} 
            className="flex justify-between items-center p-2 border rounded"
        >
            <span>{fileName}</span>
            <div className="space-x-2">
                <Button 
                    onClick={() => handleFileDownload(fileName)}
                    size="sm"
                >
                    Download
                </Button>
                <Button 
                    onClick={() => handleFileDelete(fileName)}
                    size="sm"
                    variant="destructive"
                >
                    Delete
                </Button>
            </div>
        </li>
    ))}
</ul>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

export default App;