public class ChessBoard
{
    private const int BLACK = 0;
    private const int WHITE = 1;

    private char[,] board = new char[8, 8];
    private Piece[,] piece = new Piece[2, 16];


    private void GenerateBlackPiece()
    {
        const int color = BLACK;
        
        piece[color, 0] = PieceFactory(color, Piece.PieceType.King, 4, 0);
        piece[color, 1] = PieceFactory(color, Piece.PieceType.Queen, 3, 0);
        piece[color, 2] = PieceFactory(color, Piece.PieceType.Bishop, 2, 0);
        piece[color, 3] = PieceFactory(color, Piece.PieceType.Bishop, 5, 0);
        piece[color, 4] = PieceFactory(color, Piece.PieceType.Knight, 1, 0);
        piece[color, 5] = PieceFactory(color, Piece.PieceType.Knight, 6, 0);
        piece[color, 6] = PieceFactory(color, Piece.PieceType.Rook, 0, 0);
        piece[color, 7] = PieceFactory(color, Piece.PieceType.Rook, 7, 0);

        for (int i = 8; i < 16; i++)
        {
            piece[color, i] = PieceFactory(color, Piece.PieceType.Pawn, i-8, 1);
        }
    }

    private void GenerateWhitePiece()
    {
        int color = WHITE;
        
        piece[color, 0] = PieceFactory(color, Piece.PieceType.King, 4, 7);
        piece[color, 1] = PieceFactory(color, Piece.PieceType.Queen, 3, 7);
        
        piece[color, 2] = PieceFactory(color, Piece.PieceType.Bishop, 2, 7);
        piece[color, 3] = PieceFactory(color, Piece.PieceType.Bishop, 5, 7);
        piece[color, 4] = PieceFactory(color, Piece.PieceType.Knight, 1, 7);
        piece[color, 5] = PieceFactory(color, Piece.PieceType.Knight, 6, 7);
        piece[color, 6] = PieceFactory(color, Piece.PieceType.Rook, 0, 7);
        piece[color, 7] = PieceFactory(color, Piece.PieceType.Rook, 7, 7);
        
        for (int i = 8; i < 16; i++)
        {
            piece[color, i] = PieceFactory(color, Piece.PieceType.Pawn, i-8, 6);
        }
    }

    
    public ChessBoard()
    {
        for (var i = 0; i < 8; i++)
        {
            for (var j = 0; j < 8; j++)
            {
                board[i,j] = ' ';
            }
        }
        GenerateBlackPiece();
        GenerateWhitePiece();
        
        // Assign the initial position of all pieces
        for (int color = 0; color <= 1; color++)
        {
            for (int i = 0; i < 16; i++)
            {
                int x = piece[color, i].x;
                int y = piece[color, i].y;
                board[y, x] = piece[color, i].getType();
            }
        }
    }

    Piece PieceFactory(int color, Piece.PieceType type, int x, int y)
    {
        Piece newPiece = new Piece(); 
        newPiece.Set(color == BLACK, type, x, y);
        return newPiece;
    }
    public void print()
    {
        Console.WriteLine("    A | B | C | D | E | F | G | H\n");
        for (var i = 0; i < 8; i++)
        {
            Console.Write($"{8-i}  |");
            for (var j = 0; j < 8; j++)
            {
                Console.Write($"{board[i, j]}|");
            }
            Console.WriteLine("");
        }
    }
}