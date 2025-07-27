from flask import Flask, request, jsonify
from flask_sqlalchemy import SQLAlchemy
from datetime import datetime

app = Flask(__name__)

# SQLite database configuration
app.config['SQLALCHEMY_DATABASE_URI'] = 'sqlite:///demo.db'
app.config['SQLALCHEMY_TRACK_MODIFICATIONS'] = False

db = SQLAlchemy(app)


# --------------------------
# User Model
# --------------------------
class User(db.Model):
    id = db.Column(db.Integer, primary_key=True)
    name = db.Column(db.String(100), nullable=False)
    email = db.Column(db.String(120), nullable=False, unique=True)  # Email must be unique
    username = db.Column(db.String(80), nullable=False, unique=True)  # Username must be unique
    created_at = db.Column(db.DateTime, default=datetime.utcnow)
    updated_at = db.Column(db.DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    def to_dict(self):
        return {
            'id': self.id,
            'name': self.name,
            'email': self.email,
            'username': self.username,
            'created_at': self.created_at.isoformat(),
            'updated_at': self.updated_at.isoformat()
        }


# --------------------------
# Create tables automatically
# --------------------------
with app.app_context():
    db.create_all()


# --------------------------
# Routes
# --------------------------

# Create User
@app.route('/users', methods=['POST'])
def create_user():
    data = request.get_json()

    # Check if user with same email already exists
    existing_user_email = User.query.filter_by(email=data['email']).first()
    if existing_user_email:
        return jsonify({"error": "User with this email already exists"}), 409

    # Check if user with same username already exists
    existing_user_username = User.query.filter_by(username=data['username']).first()
    if existing_user_username:
        return jsonify({"error": "User with this username already exists"}), 409

    user = User(name=data['name'], email=data['email'], username=data['username'])
    db.session.add(user)
    db.session.commit()

    return jsonify({"message": "User created successfully", "user": user.to_dict()}), 201


# Get All Users
@app.route('/users', methods=['GET'])
def get_users():
    users = User.query.all()
    return jsonify([user.to_dict() for user in users])


# Get Single User by ID
@app.route('/users/<int:user_id>', methods=['GET'])
def get_user(user_id):
    user = User.query.get_or_404(user_id)
    return jsonify(user.to_dict())


# Update User
@app.route('/users/<int:user_id>', methods=['PUT'])
def update_user(user_id):
    user = User.query.get_or_404(user_id)
    data = request.get_json()

    user.name = data['name']
    user.email = data['email']
    user.username = data['username']
    user.updated_at = datetime.utcnow()

    db.session.commit()
    return jsonify({"message": "User updated successfully", "user": user.to_dict()})


# Delete User
@app.route('/users/<int:user_id>', methods=['DELETE'])
def delete_user(user_id):
    user = User.query.get_or_404(user_id)
    db.session.delete(user)
    db.session.commit()
    return jsonify({"message": "User deleted successfully"})


# --------------------------
# Run Flask App
# --------------------------
if __name__ == '__main__':
    app.run(debug=True)
